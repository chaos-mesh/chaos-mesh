# Overview

Chaos Mesh is a meritocratic, consensus-based community project. Anyone with an interest in the project can join the community, contribute to the project design and participate in the decision making process. This document describes how that participation takes place and how to set about earning merit within the project community.

# Roles and responsibilities

## Users

Users are community members who have a need for the project. They are the most important members of the community, without them, the project would have no purpose. Anyone can be a user, there are no special requirements.

The project asks its users to participate in the project and community as much as possible. User contributions enable the project team to ensure that they are satisfying the needs of those users. Common user contributions include (but not limited to):

- Evangelising about the project (e.g. a link on a website or word-of-mouth awareness raising).
- Supporting others.
- Informing developers of strengths and weaknesses from a new user’s perspective.
- Providing moral support (a ‘thank you’ goes a long way).
- Providing financial support (the software is open source, but its developers need to eat).

Users who continue to engage with the project may either be recognized as Chaos Mesh Evangelists or find themselves becoming contributors, as described in the following section.

## Contributor

Contributors are community members who contribute in concrete ways to the project. Anyone can contribute to the project and become a contributor, regardless of their skills. There is no expectation of commitment to the project, no specific skill requirements and no selection process. There are many ways to contribute to the project, which may be one or more of the following (but not limited to):

- Reporting or fixing bugs.
- Identifying requirements.
- Improving the Chaos Mesh website.
- Assisting with project infrastructure.
- Writing documentation.
- Adding features.
- Answering questions on social platforms such as Twitter, community Slack channel.

For details, see [Contributing to Chaos Mesh](https://github.com/chaos-mesh/chaos-mesh/blob/master/CONTRIBUTING.md). For first time contributors, the community Slack channel is the most appropriate place to ask for help.

As one gains experience and familiarity with the project and as their commitment to the community increases, they may find themselves being nominated for committership at some stage.

## Committer

Committers are active community members who have shown that they are committed to the continuous development of the project through ongoing engagement with the community. Committership allows contributors to more easily carry on with their project-related activities by giving them direct access to the project’s resources.

Typically, a potential committer will need to show that they have sufficient understanding of the project, its objectives and its strategy. To become a committer, you are expected to:

- Express interest to the existing maintainers that you are interested in becoming a committer.
- Have contributed 6 substantial PRs or above.
- Have above average understanding of the project, its goals and directions.

Contributors that meet the above requirements will be nominated by an existing maintainer to become a committer. The existing maintainers will confer and decide whether to grant committer status or not.

Committers are expected to review issues and PRs. Their LGTM counts towards the required LGTM count to merge a PR. While committership indicates a valued member of the community who has demonstrated a healthy respect for the project’s aims and objectives, their work continues to be reviewed by the community before acceptance in an official release.

A committer who shows an above-average level of contribution to the project, particularly with respect to its strategic direction and long-term health, may be nominated to become a maintainer. This role is described below.

## Maintainer

Maintainers are first and foremost committers that have shown they are committed to the long term success of a project. They are the planners and designers of the Chaos Mesh project. Maintainership is about building trust with the current maintainers of the project and being a person that they can depend on to make decisions in the best interest of the project in a consistent manner.

Committers wanting to become maintainers are expected to:

- Enable adoptions or ecosystems.
- Collaborate well.
- Demonstrate a deep and comprehensive understanding of Chaos Mesh's architecture, technical goals, and directions.
- Actively engage with major Chaos Mesh feature proposals and implementations.

A new maintainer must be nominated by an existing maintainer. The nominating maintainer will create a PR to update the [Maintainers List](https://github.com/chaos-mesh/chaos-mesh/blob/master/MAINTAINERS.md). Upon consensus of incumbent maintainers, the PR will be approved and the new maintainer becomes active.

If a maintainer is no longer interested or cannot perform the maintainer duties listed above, they should volunteer to be moved to emeritus status. In extreme cases this can also occur by a vote of the maintainers per the voting process, as mentioned below.

# Approving PRs

PRs may be merged only after receiving at least two approvals (LGTMs) from committers or maintainers. However, maintainers can sidestep this rule under justifiable circumstances. For example:

- If a CI tool is broken, may override the tool to still submit the change.
- Minor typos or fixes for broken tests.
- The change was approved through other means than the standard process.

# Decision Making Process

Ideally, all project decisions are resolved by consensus via a PR or GitHub issue. Any of the day-to-day project maintenance can be done by a [lazy consensus model(https://communitymgt.fandom.com/wiki/Lazy_consensus).

Community or project level decisions such as RFC submission, creating a new project, maintainer promotion, and major updates on GOVERNANCE must be brought to broader awareness of the community via community meetings, GitHub discussions,  and slack channels. A supermajority (2/3) approval from Maintainers is required for such approvals.

In general, we prefer that technical issues and maintainer membership are amicably worked out between the persons involved. If a dispute cannot be decided independently, the maintainers can be called in to resolve the issue by voting. For voting, a specific statement of what is being voted on should be added to the relevant github issue or PR, and a link to that issue or PR added to the maintainers meeting agenda document. Maintainers should indicate their yes/no vote on that issue or PR, and after a suitable period of time, the votes will be tallied and the outcome noted.

Decision making must comply with the [CNCF Code of Conduct](https://github.com/chaos-mesh/chaos-mesh/blob/master/CODE_OF_CONDUCT.md).

## Proposal process

We use a Request for Comments (RFC) process for any substantial changes to Chaos Mesh. This process involves an upfront design that will provide increased visibility to the community. If you're considering a PR that will bring in a new feature that may affect how Chaos Mesh is implemented, or may be a breaking change, then you should start with a RFC. The process is documented in the [RFC repository](https://github.com/chaos-mesh/rfcs)) and has [a template](https://github.com/chaos-mesh/rfcs/blob/main/template.md) for you to get started.