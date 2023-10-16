package aws

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsNetworkACLRuleResourceType = "aws_network_acl_rule"

var protocolsNumbers = map[string]int{
	// defined at https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml
	"all":             -1,
	"hopopt":          0,
	"icmp":            1,
	"igmp":            2,
	"ggp":             3,
	"ipv4":            4,
	"st":              5,
	"tcp":             6,
	"cbt":             7,
	"egp":             8,
	"igp":             9,
	"bbn-rcc-mon":     10,
	"nvp-ii":          11,
	"pup":             12,
	"argus":           13,
	"emcon":           14,
	"xnet":            15,
	"chaos":           16,
	"udp":             17,
	"mux":             18,
	"dcn-meas":        19,
	"hmp":             20,
	"prm":             21,
	"xns-idp":         22,
	"trunk-1":         23,
	"trunk-2":         24,
	"leaf-1":          25,
	"leaf-2":          26,
	"rdp":             27,
	"irtp":            28,
	"iso-tp4":         29,
	"netblt":          30,
	"mfe-nsp":         31,
	"merit-inp":       32,
	"dccp":            33,
	"3pc":             34,
	"idpr":            35,
	"xtp":             36,
	"ddp":             37,
	"idpr-cmtp":       38,
	"tp++":            39,
	"il":              40,
	"ipv6":            41,
	"sdrp":            42,
	"ipv6-route":      43,
	"ipv6-frag":       44,
	"idrp":            45,
	"rsvp":            46,
	"gre":             47,
	"dsr":             48,
	"bna":             49,
	"esp":             50,
	"ah":              51,
	"i-nlsp":          52,
	"swipe":           53,
	"narp":            54,
	"mobile":          55,
	"tlsp":            56,
	"ipv6-icmp":       58,
	"ipv6-nonxt":      59,
	"ipv6-opts":       60,
	"61":              61,
	"cftp":            62,
	"63":              63,
	"sat-expak":       64,
	"kryptolan":       65,
	"rvd":             66,
	"ippc":            67,
	"68":              68,
	"sat-mon":         69,
	"visa":            70,
	"ipcv":            71,
	"cpnx":            72,
	"cphb":            73,
	"wsn":             74,
	"pvp":             75,
	"br-sat-mon":      76,
	"sun-nd":          77,
	"wb-mon":          78,
	"wb-expak":        79,
	"iso-ip":          80,
	"vmtp":            81,
	"secure-vmtp":     82,
	"vines":           83,
	"ttp":             84,
	"nsfnet-igp":      85,
	"dgp":             86,
	"tcf":             87,
	"eigrp":           88,
	"ospfigp":         89,
	"sprite-rpc":      90,
	"larp":            91,
	"mtp":             92,
	"ax.25":           93,
	"ipip":            94,
	"micp":            95,
	"scc-sp":          96,
	"etherip":         97,
	"encap":           98,
	"99":              99,
	"gmtp":            100,
	"ifmp":            101,
	"pnni":            102,
	"pim":             103,
	"aris":            104,
	"scps":            105,
	"qnx":             106,
	"a/n":             107,
	"ipcomp":          108,
	"snp":             109,
	"compaq-peer":     110,
	"ipx-in-ip":       111,
	"vrrp":            112,
	"pgm":             113,
	"114":             114,
	"l2tp":            115,
	"dd":              116,
	"iatp":            117,
	"stp":             118,
	"srp":             119,
	"uti":             120,
	"smp":             121,
	"sm":              122,
	"ptp":             123,
	"isis-over-ipv4":  124,
	"fire":            125,
	"crtp":            126,
	"crudp":           127,
	"sscopmce":        128,
	"iplt":            129,
	"sps":             130,
	"pipe":            131,
	"sctp":            132,
	"fc":              133,
	"rsvp-e2e-ignore": 134,
	"mobility-header": 135,
	"udplite":         136,
	"mpls-in-ip":      137,
	"manet":           138,
	"hip":             139,
	"shim6":           140,
	"wesp":            141,
	"rohc":            142,
	"253":             253,
	"254":             254,
}

func initAwsNetworkACLRuleMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsNetworkACLRuleResourceType, func(res *resource.Resource) {
		res.Attrs.DeleteIfDefault("icmp_code")
		res.Attrs.DeleteIfDefault("icmp_type")

		// Since it seems that AWS only works with protocol number, we should normalize when we got a protocol string
		// and transform it to its proper protocol number
		// We iterate on ingress and egresses to modify protocols that are full string like "tcp" to "6"
		//
		// References:
		// - https://github.com/hashicorp/terraform-provider-aws/blob/1194e7a11e6b74f1f4834c90940ffef0f6557982/aws/network_acl_entry.go#L69
		proto := res.Attrs.GetString("protocol")
		if number, isNotProtoAsNumber := protocolsNumbers[*proto]; isNotProtoAsNumber {
			_ = res.Attrs.SafeSet([]string{"protocol"}, strconv.Itoa(number))
		}

		// For some reason, when deserialising the state, this field is deserialized as a float
		// We need to make this homogeneous between remote and IaC so we cast this to an int64
		// The real type returned by AWS SDK is int64
		ruleNumber := (*res.Attrs)["rule_number"]
		if v, isFloat := ruleNumber.(float64); isFloat {
			_ = res.Attrs.SafeSet([]string{"rule_number"}, int64(v))
		}

		// ID can be different even if the resource is the same.
		// protocol is taken into account while creating the ID, if you set protocol="tcp" you'll end with
		// a resource with a different ID than if you set protocol="6" which is the same
		// To be able to match resources, we rewrite ID to always use protocol as a number (we just normalized this above)
		//
		// While reading remote we always got protocol as a number.
		// We cannot predict how the user decided to write the protocol on IaC side.
		// This workaround is mandatory to harmonize resources ID
		res.Id = CreateNetworkACLRuleID(
			*res.Attrs.GetString("network_acl_id"),
			(*res.Attrs)["rule_number"].(int64),
			*res.Attrs.GetBool("egress"),
			*res.Attrs.GetString("protocol"),
		)
		_ = res.Attrs.SafeSet([]string{"id"}, res.Id)

		res.Attrs.DeleteIfDefault("cidr_block")
		res.Attrs.DeleteIfDefault("ipv6_cidr_block")
	})
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AwsNetworkACLRuleResourceType, func(res *resource.Resource) map[string]string {

		ruleNumber := strconv.FormatInt((*res.Attrs)["rule_number"].(int64), 10)
		if ruleNumber == "32767" {
			ruleNumber = "*"
		}

		attrs := map[string]string{
			"Network":     *res.Attrs.GetString("network_acl_id"),
			"Egress":      strconv.FormatBool(*res.Attrs.GetBool("egress")),
			"Rule number": ruleNumber,
		}

		if proto := res.Attrs.GetString("protocol"); proto != nil {
			if *proto == "-1" {
				*proto = "All"
			}
			attrs["Protocol"] = *proto
		}

		if res.Attrs.GetFloat64("from_port") != nil && res.Attrs.GetFloat64("to_port") != nil {
			attrs["Port range"] = fmt.Sprintf("%d - %d",
				int64(*res.Attrs.GetFloat64("from_port")),
				int64(*res.Attrs.GetFloat64("to_port")),
			)
		}

		if cidr := res.Attrs.GetString("cidr_block"); cidr != nil && *cidr != "" {
			attrs["CIDR"] = *cidr
		}

		if cidr := res.Attrs.GetString("ipv6_cidr_block"); cidr != nil && *cidr != "" {
			attrs["CIDR"] = *cidr
		}

		return attrs
	})
}

func CreateNetworkACLRuleID(networkAclId string, ruleNumber int64, egress bool, protocol string) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s-", networkAclId))
	buf.WriteString(fmt.Sprintf("%d-", ruleNumber))
	buf.WriteString(fmt.Sprintf("%t-", egress))
	buf.WriteString(fmt.Sprintf("%s-", protocol))
	return fmt.Sprintf("nacl-%d", hashcode.String(buf.String()))
}
