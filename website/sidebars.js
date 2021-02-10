module.exports = {
  docs: {
    Introduction: ["intro", "installation", "limitations"],
    Commands: [
      {
        Scan: ["cmd/scan/iac-source", "cmd/scan/output", "cmd/scan/filtering"],
        Flags: ["cmd/flags/error-reporting", "cmd/flags/version-check"],
        Completion: ["cmd/completion/script"],
      },
    ],
    Providers: [
      {
        AWS: ["providers/aws/authentication", "providers/aws/resources"],
      },
    ],
  },
};
