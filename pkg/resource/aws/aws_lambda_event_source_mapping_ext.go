package aws

func (r *AwsLambdaEventSourceMapping) Attributes() map[string]string {
	attrs := make(map[string]string)
	if r.EventSourceArn != nil && *r.EventSourceArn != "" && r.FunctionName != nil && *r.FunctionName != "" {
		attrs["Source"] = *r.EventSourceArn
		attrs["Dest"] = *r.FunctionName
	}
	return attrs
}
