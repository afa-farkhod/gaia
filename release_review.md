> [!NOTE]
> working repo: https://github.com/cosmos/gaia

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

- Discord Announcement
- Upgrade proposal
- v20.0.0 release details
- `git show v20.0.0` → who tagged? (mpoke)
- CHANGELOG
- comparing changes (v19.2.0 vs v20.0.0)

---

### 3️⃣ Security fix related patch

> [!NOTE]
> Binary only release
> Do not build this binary from source: the security fix is not included in the tagged commit.
> Gaia has already been privately patched by >2/3 of the voting power - the chain is secure
> Bump cometbft from v0.38.20 to v0.38.21 to address critical security vulnerability in CometBFT detailed here

- Discord Announcement
- v25.3.1 release details
- comparing changes (v25.3.0 vs v25.3.2)

---

### 4️⃣ Consensus failure related patch fix

> [!NOTE]
> bug causing the chain halt
> post-mortem
> incident occurred slightly after the scheduled v17 software upgrade took place, the upgrade triggered a bug when a validator leaves the active set of validators and another validator takes its place

```
# reference: https://discord.com/channels/669268347736686612/798937713474142229/1248000697589436437

2024-06-05 19:21:51: 7:21PM INF validator un-jailed module=x/staking validator=cosmosvalcons1eemgqv7hylr29zvq4mmnwtgjgvnusy9g4nlklf
2024-06-05 19:21:51: 7:21PM ERR CONSENSUS FAILURE!!! err="more validators than maxValidators found" module=consensus stack="goroutine 1403 
```




- Discord Announcement
- v17.2.0 release details

comparing changes (v17.1.0 vs v17.2.0)

