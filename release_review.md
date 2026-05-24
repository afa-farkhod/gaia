> [!NOTE]
> working repo: https://github.com/cosmos/gaia

---

### 1️⃣ Ordinary Upgrade Release:

> [!NOTE]
> dependencies-only point release
> no new features, no bug fixes, 
> no protocol-level logic changes inside
> only three coordinated bumps to the IBC stack

- [Discord Announcement](https://discord.com/channels/669268347736686612/1085152096380260372/1504103576044048644)
- [Upgrade proposal](https://www.mintscan.io/cosmos/proposals/1035)
- [v27.3.0](https://github.com/cosmos/gaia/releases/tag/v27.3.0) release details:
  - `git show v27.3.0` -> who tagged? (Dante)
  - [release workflow on tag push](https://github.com/cosmos/gaia/blob/5b00e24daae7f4bdf34b2d24744b8049bb8e6efa/.github/workflows/release.yml#L7)
  - [git action run](https://github.com/cosmos/gaia/actions/runs/25689480440)
- [CHANGELOG](https://github.com/cosmos/gaia/blob/v27.3.0/CHANGELOG.md)
- [upgrading guide](https://github.com/cosmos/gaia/blob/v27.3.0/UPGRADING.md)
- [comparing changes](https://github.com/cosmos/gaia/compare/v27.2.0...v27.3.0) (v27.2.0 vs v27.3.0)
  - TL;DR: dependence bot auto detects new changes, makes PR against the [release/v27.x](https://github.com/cosmos/gaia/tree/release/v27.x) branch, in the end all commits are cherry picked  
  - https://github.com/cosmos/gaia/commit/14b560830add97859012d49299c18da3852dfe5b    
  - https://github.com/cosmos/gaia/commit/71d78edb617bdf42f752bdf780be9115ce669d65 
  - https://github.com/cosmos/gaia/pull/4037 

---

### 2️⃣ Breaking changes included release:

> [!NOTE]
> major version bump focused on the permissionless ICS rollout (making consumer chains permissionless and letting validators outside the top 180 participate in cross-chain security)
> API-breaking and state-breaking

- [Discord Announcement](https://discord.com/channels/669268347736686612/1085152096380260372/1289324152343232572)
- [Upgrade proposal](https://www.mintscan.io/cosmos/proposals/966)
- [v20.0.0](https://github.com/cosmos/gaia/releases/tag/v20.0.0) release details
- `git show v20.0.0` → who tagged? (mpoke)
- [CHANGELOG](https://github.com/cosmos/gaia/blob/v27.3.0/CHANGELOG.md#v2000)
- [comparing changes](https://github.com/cosmos/gaia/compare/v19.2.0...v20.0.0) (v19.2.0 vs v20.0.0)

---

### 3️⃣ Security fix related patch

> [!NOTE]
> Binary only release
> Do not build this binary from source: the security fix is not included in the tagged commit.
> Gaia has already been privately patched by >2/3 of the voting power - the chain is secure
> Bump [cometbft](https://github.com/cometbft/cometbft) from [v0.38.20](https://github.com/cometbft/cometbft/releases/tag/v0.38.21) to v0.38.21 to address critical security vulnerability in CometBFT detailed [here](https://github.com/cometbft/cometbft/security/advisories/GHSA-c32p-wcqj-j677)

- [Discord Announcement](https://discord.com/channels/669268347736686612/1085152096380260372/1461088582415548530)
- [v25.3.1](https://github.com/cosmos/gaia/releases/tag/v25.3.1) release details
- [comparing changes](https://github.com/cosmos/gaia/compare/v25.3.0...v25.3.2) (v25.3.0 vs v25.3.2)

---

### 4️⃣ Consensus failure related patch fix

> [!NOTE]
> bug causing the chain halt
> [post-mortem](https://forum.cosmos.network/t/cosmos-hub-v17-1-chain-halt-post-mortem/13899)
> incident occurred slightly after the scheduled [v17 software upgrade](https://www.mintscan.io/cosmos/proposals/924) took place, the upgrade triggered a bug when a validator leaves the active set of validators and another validator takes its place

```
# reference: https://discord.com/channels/669268347736686612/798937713474142229/1248000697589436437

2024-06-05 19:21:51: 7:21PM INF validator un-jailed module=x/staking validator=cosmosvalcons1eemgqv7hylr29zvq4mmnwtgjgvnusy9g4nlklf
2024-06-05 19:21:51: 7:21PM ERR CONSENSUS FAILURE!!! err="more validators than maxValidators found" module=consensus stack="goroutine 1403 
```

#### [Timeline](https://forum.cosmos.network/t/cosmos-hub-v17-1-chain-halt-post-mortem/13899#p-31338-timeline-2)

| Event | Block Height | Time (UTC) |
|---|---|---|
| Chain upgrade started | [20739800](https://www.mintscan.io/cosmos/block/20739800) | 16:58 (on June 5th, 2024) |
| Chain upgrade completed | [20739802](https://www.mintscan.io/cosmos/block/20739802) | 17:15 |
| Chain halts | [20740970](https://www.mintscan.io/cosmos/block/20740970) | 19:21 |
| Informed (over Slack) by the Informal Staking team that the chain has halted | chain is halted | 19:46 |
| Hypha (over Slack) confirms the error in their mainnet node as well | chain is halted | 19:47 |
| Hypha, the Informal Systems Cosmos Hub team, and Binary Builders meet on Zoom to fix the issue | chain is halted | 20:06 |
| Fix the issue and cut Cosmos SDK [v0.47.15-ics-lsm](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.47.15-ics-lsm) | chain is halted | 21:35 (from `git show v0.47.15-ics-lsm`) |
| Cut Gaia [v17.2.0](https://github.com/cosmos/gaia/releases/tag/v17.2.0) (using Cosmos SDK `v0.47.15-ics-lsm`) | chain is halted | 22:28 (from `git show v17.2.0`) |
| Published the release binaries after running automated tests | chain is halted | 22:57 (see the [action](https://github.com/cosmos/gaia/actions/runs/9392082523)) |
| Chain resumes | [20740972](https://www.mintscan.io/cosmos/block/20740972) | 0:02 (on June 6th, 2024) |

The total time the chain was halted was around **4 hours and 40 minutes**.

- [Discord Announcement](https://discord.com/channels/669268347736686612/1085152096380260372/1248009908159516812)
- [v17.2.0](https://github.com/cosmos/gaia/releases/tag/v17.2.0) release details
- [comparing changes](https://github.com/cosmos/gaia/compare/v17.1.0...v17.2.0) (v17.1.0 vs v17.2.0)
- [comparing changes](https://github.com/cosmos/gaia/compare/v17.1.0...v17.2.0) (v17.1.0 vs v17.2.0)

