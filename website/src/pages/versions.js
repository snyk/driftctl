import React from "react";
import useDocusaurusContext from "@docusaurus/useDocusaurusContext";
import Link from "@docusaurus/Link";
import Layout from "@theme/Layout";
import { useVersions, useLatestVersion } from "@theme/hooks/useDocs";

const Version = () => {
  const { siteConfig } = useDocusaurusContext();
  const versions = useVersions();
  const latestVersion = useLatestVersion();
  const currentVersion = versions.find((version) => version.name === "current");
  const pastVersions = versions.filter(
    (version) => version !== latestVersion && version.name !== "current"
  );
  const repoUrl = `https://github.com/${siteConfig.organizationName}/${siteConfig.projectName}`;

  return (
    <Layout title="Versions" description="Documentation versions for driftctl">
      <main className="container margin-vert--lg">
        <h1>Documentation versions for driftctl</h1>

        {latestVersion && (
          <div className="margin-bottom--lg">
            <h3 id="latest">Current version (Latest)</h3>
            <p>
              Here you can find the documentation for current released version.
            </p>
            <table>
              <tbody>
                <tr>
                  <th>{latestVersion.name}</th>
                  <td>
                    <Link to={latestVersion.path}>Documentation</Link>
                  </td>
                  <td>
                    <a href={`${repoUrl}/releases/tag/v${latestVersion.name}`}>
                      Release Notes
                    </a>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        )}

        <div className="margin-bottom--lg">
          <h3 id="next">Next version (Unreleased)</h3>
          <p>
            Here you can find the documentation for work-in-progress unreleased
            version.
          </p>
          <table>
            <tbody>
              <tr>
                <th>{currentVersion.label}</th>
                <td>
                  <Link to={currentVersion.path}>Documentation</Link>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        {pastVersions.length > 0 && (
          <div className="margin-bottom--lg">
            <h3 id="archive">Past versions</h3>
            <p>
              Here you can find documentation for previous versions of driftctl.
            </p>
            <table>
              <tbody>
                {pastVersions.map((version) => (
                  <tr key={version.name}>
                    <th>{version.label}</th>
                    <td>
                      <Link to={version.path}>Documentation</Link>
                    </td>
                    <td>
                      <a href={`${repoUrl}/releases/tag/v${version.name}`}>
                        Release Notes
                      </a>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </main>
    </Layout>
  );
};

export default Version;
