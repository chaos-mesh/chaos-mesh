# Chaos Mesh Security Policy

Chaos Mesh is a growing community devoted to building a chaos engineering ecology under the cloud-native realm. As maintainers of Chaos Mesh, we attach great importance to code security. We are very grateful to the users, security vulnerability researchers, etc. for reporting security vulnerabilities to us. All reported security vulnerabilities will be carefully assessed, addressed, and answered by us.

## Supported Versions

We provide security updates for the two most recent minor versions released on GitHub.

For example, if `v1.2.2` is the most recent stable version, we will address security updates for `v1.1.0` and later, Once `v1.3.0` is released, we will no longer provide updates for `v1.1.x` releases.

## Reporting a Vulnerability

To report a security problem in Chaos Mesh, please contact the Chaos Mesh Security Team: chaos-mesh-security@lists.cncf.io.
The team will help diagnose the severity of the issue and determine how to address the issue. Issues deemed to be non-critical will be filed as GitHub issues. Critical issues will receive immediate attention and be fixed as quickly as possible.

## Disclosure policy

For known public security vulnerabilities, we will disclose the disclosure as soon as possible after receiving the report. Vulnerabilities discovered for the first time will be disclosed in accordance with the following process:

1. The received security vulnerability report shall be handed over to the security team for follow-up coordination and repair work.
2. After the vulnerability is confirmed, we will create a draft Security Advisory on Github that lists the details of the vulnerability.
3. Invite related personnel to discuss the fix.
4. Fork the temporary private repository on Github, and collaborate to fix the vulnerability.
5. After the fixed code is merged into all supported versions, the vulnerability will be publicly posted in the GitHub Advisory Database.
