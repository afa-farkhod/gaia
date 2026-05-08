# Cosmos Hub (Gaia) — Reference

> **Current version:** v28.0.0 (module path `github.com/cosmos/gaia/v28`)
> **Go version:** 1.25.7 (minimum required)
> **Binary name:** `gaiad`

---

## Table of Contents

1. [What Is Gaia](#1-what-is-gaia)
2. [Repository Layout](#2-repository-layout)
3. [Architecture Overview](#3-architecture-overview)
4. [Application Bootstrap (`app/`)](#4-application-bootstrap-app)
5. [Module Inventory](#5-module-inventory)
   - 5.1 [Standard Cosmos SDK Modules](#51-standard-cosmos-sdk-modules)
   - 5.2 [IBC Stack](#52-ibc-stack)
   - 5.3 [Interchain Security (ICS)](#53-interchain-security-ics)
   - 5.4 [CosmWasm](#54-cosmwasm)
   - 5.5 [Gaia-Specific Custom Modules](#55-gaia-specific-custom-modules)
6. [Ante Handler Pipeline](#6-ante-handler-pipeline)
7. [Keepers (`app/keepers/`)](#7-keepers-appkeepers)
8. [Upgrade Handlers (`app/upgrades/`)](#8-upgrade-handlers-appupgrades)
9. [Key Dependencies](#9-key-dependencies)
10. [Protobuf Management](#10-protobuf-management)
11. [Build System](#11-build-system)
12. [Testing Strategy](#12-testing-strategy)
13. [CI/CD Workflows](#13-cicd-workflows)
14. [Release Process](#14-release-process)
15. [Upgrade Process (Validator/Node)](#15-upgrade-process-validatornode)
16. [State Compatibility Rules](#16-state-compatibility-rules)
17. [Version Bump Procedure](#17-version-bump-procedure)
18. [Contributing Workflow](#18-contributing-workflow)
19. [Local Development Quick Reference](#19-local-development-quick-reference)

---

## 1. What Is Gaia

Gaia is the implementation of the **Cosmos Hub** — the first and reference blockchain in the Cosmos ecosystem. It is a Proof-of-Stake chain built with the Cosmos SDK whose native token is **ATOM (uatom)**. Gaia's primary roles are:

- Running the ATOM economic zone (staking, governance, distribution).
- Acting as an **ICS Provider chain** that secures consumer chains via Interchain Security (CCV protocol).
- Routing IBC packets through Packet Forward Middleware.
- Hosting CosmWasm smart contracts.
- Supporting native liquid staking via `x/liquid`.

---

## 2. Repository Layout

```
gaia/
├── ante/                    # Custom ante decorators
├── app/
│   ├── app.go               # GaiaApp struct + wiring
│   ├── config.go            # Bech32/address codec config
│   ├── const.go             # Chain constants
│   ├── export.go            # State export for genesis
│   ├── genesis.go           # Genesis init/load
│   ├── genesis_account.go   # Vesting account support
│   ├── modules.go           # Module list, maccPerms, BeginBlock/EndBlock order
│   ├── post.go              # PostHandler wiring
│   ├── helpers/             # Test helpers
│   ├── keepers/
│   │   ├── keepers.go       # All keeper instantiation + IBC router wiring
│   │   └── keys.go          # Store keys
│   ├── params/              # EncodingConfig helpers
│   ├── sim/                 # Simulation helpers
│   └── upgrades/
│       └── v28_0_0/
│           ├── constants.go # Upgrade name constant
│           └── upgrades.go  # Upgrade handler logic
├── cmd/
│   └── gaiad/
│       ├── main.go          # Entry point
│       └── cmd/
│           ├── root.go      # Root cobra command
│           ├── genaccounts.go
│           ├── testnet.go
│           └── bech32_convert.go
├── contrib/
│   ├── devtools/Makefile    # Tool targets included in main Makefile
│   ├── images/              # Docker images for localnet
│   ├── scripts/             # Upgrade test scripts
│   └── testnets/            # Remote/local testnet configs
├── docs/                    # Docusaurus site
├── pkg/address/             # Address utilities
├── proto/
│   └── gaia/
│       ├── liquid/          # x/liquid proto definitions
│       └── metaprotocols/   # x/metaprotocols proto definitions
├── tests/
│   ├── e2e/                 # Docker-based end-to-end tests
│   ├── integration/         # In-process integration tests
│   └── interchain/          # Multi-chain interchain tests (separate go.mod)
├── types/errors/            # Shared error codes
├── x/
│   ├── bank/                # Extended bank module (multisend gas surcharge)
│   ├── gov/                 # Extended governance (ICA + wasm vote validation)
│   ├── liquid/              # Native liquid staking module
│   └── metaprotocols/       # Transaction extension data support
├── go.mod
├── go.sum
├── Makefile
├── sims.mk
├── Dockerfile
├── docker-compose.yml
├── .goreleaser.yml
├── .golangci.yml
├── .mergify.yml
├── .github/workflows/       # All CI workflow definitions
└── CHANGELOG.md
```

---

## 3. Architecture Overview

Gaia follows the standard **Cosmos SDK ABCI application** pattern:

```
CometBFT consensus
       │
       ▼
  BaseApp (SDK)
       │
  ┌────┴────────────────┐
  │  AnteHandler chain  │  ← custom: fee market, wasm, IBC, gov vote, provider
  └────────────────────┘
       │
  ┌────┴────────────────┐
  │   Module Manager    │
  │  (BeginBlock order) │
  │  …all modules…      │
  │  (EndBlock order)   │
  └────────────────────┘
       │
  ┌────┴────────────────┐
  │  MsgServiceRouter   │  ← routes Msgs to module MsgServers
  └────────────────────┘
       │
  ┌────┴────────────────┐
  │  PostHandler chain  │  ← feemarket post handler
  └────────────────────┘
```

**Key design principles:**
- `GaiaApp` embeds `*baseapp.BaseApp` and `keepers.AppKeepers` (all keepers).
- Module order in `appModules()` (`app/modules.go`) controls `BeginBlock`/`EndBlock` execution order.
- The `app/keepers/keepers.go` file is the single source of truth for all keeper wiring and IBC port routing.
- IBC routing uses `porttypes.Router` built inside `NewAppKeepers`.
- Gaia **overrides** `staking` and `genutil` with ICS-aware variants (`no_valupdates_staking`, `no_valupdates_genutil`) to prevent validator set updates from bypassing ICS.

---

## 4. Application Bootstrap (`app/`)

### `GaiaApp` struct (app/app.go)

```go
type GaiaApp struct {
    *baseapp.BaseApp
    keepers.AppKeepers          // all keepers embedded
    legacyAmino       *codec.LegacyAmino
    appCodec          codec.Codec
    txConfig          client.TxConfig
    interfaceRegistry types.InterfaceRegistry
    mm                *module.Manager
    sm                *module.SimulationManager
    configurator      module.Configurator
}
```

### Registered upgrades

```go
var Upgrades = []upgrades.Upgrade{v280.Upgrade}
```

Each new major version adds a new entry. The upgrade is registered in `app.go`'s `registerUpgradeHandlers()`.

### Module permissions map (`maccPerms`)

Key entries (from `app/modules.go`):

| Module | Permissions |
|--------|-------------|
| `mint` | Minter |
| `bonded_tokens_pool` | Burner, Staking |
| `not_bonded_tokens_pool` | Burner, Staking |
| `gov` | Burner |
| `transfer` | Minter, Burner |
| `wasm` | Burner |
| `tokenfactory` | Minter, Burner |

---

## 5. Module Inventory

### 5.1 Standard Cosmos SDK Modules

| Module | Package | Notes |
|--------|---------|-------|
| `auth` | `cosmos-sdk/x/auth` | Account management, signatures |
| `authz` | `cosmos-sdk/x/authz` | Message authorization grants |
| `bank` | `gaia/x/bank` | **Custom override** — adds multisend gas surcharge |
| `consensus` | `cosmos-sdk/x/consensus` | CometBFT consensus params |
| `distribution` | `cosmos-sdk/x/distribution` | Staking reward distribution |
| `evidence` | `cosmossdk.io/x/evidence` | Double-sign evidence handling |
| `feegrant` | `cosmossdk.io/x/feegrant` | Fee allowances |
| `genutil` | `interchain-security/x/ccv/no_valupdates_genutil` | ICS-aware genesis util |
| `gov` | `gaia/x/gov` | **Custom override** — ICA + wasm vote stake validation |
| `mint` | `cosmos-sdk/x/mint` | ATOM inflation |
| `params` | `cosmos-sdk/x/params` | Legacy param subspace (deprecated, kept for migration) |
| `slashing` | `cosmos-sdk/x/slashing` | Validator slashing |
| `staking` | `interchain-security/x/ccv/no_valupdates_staking` | ICS-aware staking |
| `upgrade` | `cosmossdk.io/x/upgrade` | On-chain upgrade coordination |
| `vesting` | `cosmos-sdk/x/auth/vesting` | Vesting account types |

### 5.2 IBC Stack

| Component | Package | Notes |
|-----------|---------|-------|
| IBC core | `ibc-go/v10/modules/core` | Channel, port, connection, client |
| IBC transfer | `ibc-go/v10/modules/apps/transfer` | ICS-20 token transfers |
| ICA host | `ibc-go/v10/.../27-interchain-accounts/host` | Receive ICA txs |
| ICA controller | `ibc-go/v10/.../27-interchain-accounts/controller` | Send ICA txs |
| IBC callbacks | `ibc-go/v10/modules/apps/callbacks` | Callback middleware |
| Wasm light client | `ibc-go/modules/light-clients/08-wasm/v10` | WASM-based light clients (for ICS) |
| Tendermint light client | `ibc-go/v10/modules/light-clients/07-tendermint` | Standard Tendermint IBC client |
| PFM | `ibc-apps/middleware/packet-forward-middleware/v10` | Multi-hop IBC routing |
| Rate limiting | `ibc-apps/modules/rate-limiting/v10` | IBC transfer rate limits |

### 5.3 Interchain Security (ICS)

- **Package:** `cosmos/interchain-security/v7`
- **Role:** Gaia is the **provider chain** — it distributes its validator set to consumer chains via the CCV (Cross-Chain Validation) protocol.
- **Key keeper:** `ProviderKeeper` (type `icsproviderkeeper.Keeper`)
- The `ProviderModule` (`icsprovider.AppModule`) handles:
  - Consumer chain onboarding/offboarding
  - Validator set updates forwarded to consumer chains via IBC
  - Consumer reward distribution
- **Ante override:** `NewProviderDecorator` blocks `MsgCreateConsumer` (new consumer creation disabled in v27+).

### 5.4 CosmWasm

- **Packages:** `CosmWasm/wasmd v0.60.6` + `CosmWasm/wasmvm/v2 v2.3.2`
- Smart contract execution via `WasmKeeper`.
- Gaia Gov override validates that wasm contract votes come from staked accounts.
- `WasmClientKeeper` powers IBC-08-wasm light clients.
- Token factory bindings allow wasm contracts to mint/burn token factory tokens.

### 5.5 Gaia-Specific Custom Modules

#### `x/liquid` — Native Liquid Staking

**Purpose:** Enables tokenization of staking delegations into tradeable, on-chain liquid staking tokens.

**Key concepts:**
- **Tokenize shares:** Convert delegation → liquid tokens with denom `{validatorAddr}/{recordId}`.
- **Redeem tokens:** Burn liquid tokens → restore delegation.
- **Tokenize share record:** On-chain record associating a module account with a validator delegation.
- **Locks:** Per-address opt-out system (`UNLOCKED`, `LOCKED`, `LOCK_EXPIRING`).
- **Caps:** `GlobalLiquidStakingCap` (% of total stake) and `ValidatorLiquidStakingCap` (% per validator).

**Messages:**
| Message | Description |
|---------|-------------|
| `MsgTokenizeShares` | Convert delegation to liquid tokens |
| `MsgRedeemTokensForShares` | Burn liquid tokens, restore delegation |
| `MsgTransferTokenizeShareRecord` | Transfer reward ownership of a record |
| `MsgEnableTokenizeShares` | Begin unbonding-period countdown to re-enable tokenization |
| `MsgDisableTokenizeShares` | Lock account to prevent tokenization |
| `MsgWithdrawTokenizeShareRecordReward` | Withdraw staking rewards for one record |
| `MsgWithdrawAllTokenizeShareRecordReward` | Withdraw rewards for all owned records |
| `MsgUpdateParams` | Governance-gated param update |

**State keys:** `0x51` (params), `0x5` (TotalLiquidStakedTokens), `PendingTokenizeShareAuthorizations` queue.

**BeginBlock:** Prunes expired tokenize share locks from the queue.

**Default params:**
- `GlobalLiquidStakingCap`: 25%
- `ValidatorLiquidStakingCap`: 50%

---

#### `x/metaprotocols` — Transaction Extension Data

**Purpose:** Allows attaching arbitrary protocol-specific data to transactions via `extension_options` / `non_critical_extension_options`.

**Usage:** The chain validates the type is `ExtensionData` and that it is deserializable. The data is **not used** by the chain but is stored in the block.

```json
{
  "@type": "/gaia.metaprotocols.ExtensionData",
  "protocol_id": "my-protocol",
  "protocol_version": "1",
  "data": "<base64>"
}
```

---

#### `x/gov` — Extended Governance

**Overrides SDK's `gov` module** to add:
- Stake validation for `MsgVote` and `MsgVoteWeighted` submitted via **Interchain Accounts** (ICA) — requires the ICA's controlling account to have staked tokens.
- Same validation for votes submitted via **CosmWasm contracts** (`ante/wasm_gov_vote.go`).

---

#### `x/bank` — Extended Bank

**Overrides SDK's `bank` module** to add:
- Gas surcharge on `MsgMultiSend` to mitigate spam (added in v27).

---

## 6. Ante Handler Pipeline

Defined in `ante/ante.go:NewAnteHandler`. Decorators execute in order:

```
1.  SetUpContextDecorator          — sets gas meter, block gas limit
2.  LimitSimulationGasDecorator    — (wasm) caps gas during simulation
3.  CountTXDecorator               — (wasm) tracks tx count per block for fee calc
4.  ExtensionOptionsDecorator      — validates non_critical_extension_options types
5.  ValidateBasicDecorator         — calls ValidateBasic() on all msgs
6.  TxTimeoutHeightDecorator       — rejects txs past their timeout_height
7.  ValidateMemoDecorator          — checks memo length vs params
8.  ConsumeGasForTxSizeDecorator   — charges gas proportional to tx byte size
9.  GovVoteDecorator               — validates ICA/wasm governance votes have stake
10. ProviderDecorator              — blocks MsgCreateConsumer (ICS)
11. SetPubKeyDecorator             — links pubkey to account if not yet stored
12. ValidateSigCountDecorator      — rejects txs with too many signers
13. SigGasConsumeDecorator         — charges gas per signature
14. SigVerificationDecorator       — verifies signatures
15. IncrementSequenceDecorator     — bumps account sequence
16. RedundantRelayDecorator        — (IBC) rejects duplicate relay packets
17. FeeMarketCheckDecorator        — (Skip MEV) EIP-1559-style fee enforcement
     └─ wraps DeductFeeDecorator
```

`UseFeeMarketDecorator` flag (default `true`) can disable the FeeMarket decorator for integration testing.

---

## 7. Keepers (`app/keepers/`)

`AppKeepers` (embedded in `GaiaApp`) holds every module keeper. Key keepers:

| Field | Type | Purpose |
|-------|------|---------|
| `AccountKeeper` | `authkeeper.AccountKeeper` | Account lifecycle |
| `BankKeeper` | `bankkeeper.Keeper` | Token transfers |
| `StakingKeeper` | `*stakingkeeper.Keeper` | Validator/delegation management |
| `SlashingKeeper` | `slashingkeeper.Keeper` | Slash events |
| `DistrKeeper` | `distrkeeper.Keeper` | Reward distribution |
| `GovKeeper` | `*govkeeper.Keeper` | On-chain governance |
| `MintKeeper` | `mintkeeper.Keeper` | Inflation minting |
| `UpgradeKeeper` | `*upgradekeeper.Keeper` | Upgrade plan scheduling |
| `EvidenceKeeper` | `evidencekeeper.Keeper` | Double-sign evidence |
| `FeeGrantKeeper` | `feegrantkeeper.Keeper` | Fee allowances |
| `AuthzKeeper` | `authzkeeper.Keeper` | Message grants |
| `IBCKeeper` | `*ibckeeper.Keeper` | IBC core |
| `IBCTransferKeeper` | `ibctransferkeeper.Keeper` | ICS-20 |
| `ICAHostKeeper` | `icahostkeeper.Keeper` | ICA host |
| `ICAControllerKeeper` | `icacontrollerkeeper.Keeper` | ICA controller |
| `WasmClientKeeper` | `ibcwasmkeeper.Keeper` | Wasm IBC light clients |
| `PFMRouterKeeper` | `*pfmrouterkeeper.Keeper` | Packet forwarding |
| `RateLimitKeeper` | `ratelimitkeeper.Keeper` | IBC rate limiting |
| `ProviderKeeper` | `icsproviderkeeper.Keeper` | ICS provider |
| `WasmKeeper` | `wasmkeeper.Keeper` | CosmWasm |
| `LiquidKeeper` | `liquidkeeper.Keeper` | Native liquid staking |
| `FeeMarketKeeper` | `*feemarketkeeper.Keeper` | EIP-1559 fee market |
| `TokenFactoryKeeper` | `tokenfactorykeeper.Keeper` | Token factory |
| `ConsensusParamsKeeper` | `consensusparamkeeper.Keeper` | CometBFT params |
| `ParamsKeeper` | `paramskeeper.Keeper` | Legacy param subspaces |

---

## 8. Upgrade Handlers (`app/upgrades/`)

Each major version gets a directory `app/upgrades/vX_0_0/` with:
- `constants.go` — defines the upgrade name string (e.g., `"v28.0.0"`).
- `upgrades.go` — defines `CreateUpgradeHandler(mm, configurator, keepers)` which returns an `upgradetypes.UpgradeHandler`.

The typical pattern:
```go
func CreateUpgradeHandler(...) upgradetypes.UpgradeHandler {
    return func(c context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
        // run module migrations
        vm, err := mm.RunMigrations(ctx, configurator, vm)
        // optional: write new state, update params
        return vm, err
    }
}
```

Upgrades are registered in `app.go`:
```go
var Upgrades = []upgrades.Upgrade{v280.Upgrade}
```

When adding a new major version:
1. Create `app/upgrades/vY_0_0/constants.go` and `upgrades.go`.
2. Add the upgrade to the `Upgrades` slice.
3. Call `app.registerUpgradeHandlers()` in `NewGaiaApp`.

---

## 9. Key Dependencies

### Direct dependencies (go.mod `require` block)

| Dependency | Version | Role |
|-----------|---------|------|
| `cosmossdk.io/core` | v0.11.3 | Core SDK interfaces |
| `cosmossdk.io/store` | v1.1.2 | KV/IAVL store |
| `cosmossdk.io/x/evidence` | v0.2.0 | Evidence module |
| `cosmossdk.io/x/feegrant` | v0.2.0 | Fee grant module |
| `cosmossdk.io/x/upgrade` | v0.2.0 | Upgrade module |
| `github.com/cosmos/cosmos-sdk` | v0.53.4¹ | Core Cosmos SDK |
| `github.com/cometbft/cometbft` | v0.38.21 | BFT consensus engine |
| `github.com/cosmos/ibc-go/v10` | v10.5.0 | IBC protocol |
| `github.com/cosmos/interchain-security/v7` | v7.0.0-... | ICS provider |
| `github.com/CosmWasm/wasmd` | v0.60.6 | CosmWasm runtime |
| `github.com/CosmWasm/wasmvm/v2` | v2.3.2 | Wasm VM |
| `github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10` | v10.1.0 | PFM |
| `github.com/cosmos/ibc-apps/modules/rate-limiting/v10` | v10.1.0 | IBC rate limiting |
| `github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10` | v10.3.0 | Wasm IBC light client |
| `github.com/skip-mev/feemarket` | v1.1.2-... | EIP-1559 fee market |
| `github.com/cosmos/tokenfactory` | v0.53.5 | Token factory |
| `github.com/cosmos/gogoproto` | v1.7.2 | Protobuf encoding |
| `google.golang.org/grpc` | v1.79.3 | gRPC |
| `github.com/spf13/cobra` | v1.10.1 | CLI framework |
| `github.com/gorilla/mux` | v1.8.1 | HTTP router |
| `github.com/vektra/mockery/v2` | v2.53.4 | Mock generation |

¹ `cosmos-sdk v0.53.4` used via `replace` directive (`go.mod` declares v0.53.6 but replaces to v0.53.4 for wasmd compatibility).

### Replace directives

```
github.com/99designs/keyring      → github.com/cosmos/keyring v1.2.0
github.com/cosmos/cosmos-sdk      → github.com/cosmos/cosmos-sdk v0.53.4
github.com/dgrijalva/jwt-go       → github.com/golang-jwt/jwt/v4 v4.4.2
github.com/gin-gonic/gin          → github.com/gin-gonic/gin v1.8.1
github.com/syndtr/goleveldb       → (pinned patch version)
```

---

## 10. Protobuf Management

Gaia-specific proto files live in `proto/gaia/`:
- `proto/gaia/liquid/v1beta1/` — `x/liquid` types, msgs, queries
- `proto/gaia/metaprotocols/v1/` — `ExtensionData` type

**Toolchain:** `ghcr.io/cosmos/proto-builder:0.15.2` Docker image.

**Buf workspace:** `buf.work.yaml` defines the workspace; `proto/scripts/protocgen.sh` generates Go code.

```bash
# Useful proto Makefile targets
make proto-gen          # generate Go from .proto
make proto-swagger-gen  # generate Swagger/OpenAPI
make proto-format       # clang-format all .proto files
make proto-lint         # buf lint
make proto-check-breaking  # buf breaking check vs main branch
make proto-update-deps  # buf mod update
```

Generated files end in `.pb.go`, `.pb.gw.go`, or `.pulsar.go` — **never edit these by hand**.

---

## 11. Build System

### Prerequisites

- Go ≥ 1.25 (enforced by Makefile `check_version`)
- `gcc` (for CGO / Ledger support)
- Docker (for protobuf generation, E2E tests, static builds)

### Common Makefile targets

| Target | Description |
|--------|-------------|
| `make install` | Build and install `gaiad` to `$GOPATH/bin` |
| `make build` | Build `gaiad` to `./build/gaiad` |
| `make clean` | Remove `./build/` and `artifacts/` |
| `make lint` | Run golangci-lint v2.6.2 |
| `make lint-fix` | Run golangci-lint with `--fix` |
| `make format` | Run gofumpt + golangci-lint format |
| `make test-unit` | Unit tests (no race, 5m timeout) |
| `make test-unit-cover` | Unit tests with coverage report |
| `make test-race` | Unit tests with race detector |
| `make test-e2e` | E2E tests (35m timeout) |
| `make gen-mocks` | Regenerate mocks via mockery |
| `make proto-gen` | Regenerate protobuf Go code |
| `make start-localnet-ci` | Single-node local chain for quick testing |
| `make localnet-start` | 4-node Docker localnet |
| `make localnet-stop` | Stop 4-node Docker localnet |
| `make build-static-linux-amd64` | Static `linux/amd64` binary via Docker buildx |
| `make build-static-linux-arm64` | Static `linux/arm64` binary via Docker buildx |
| `make create-release TAG=vX.Y.Z` | Tag and push release |
| `make create-release-dry-run TAG=vX.Y.Z` | Dry-run goreleaser |
| `make bump-version OLD_VERSION=vA.B.C NEW_VERSION=vX.Y.Z` | Automated version bump |
| `make vulncheck` | Run `govulncheck` on all packages |
| `make draw-deps` | Generate dependency graph PNG |

### Build tags

| Tag | Activated when |
|-----|---------------|
| `netgo` | Always |
| `ledger` | `LEDGER_ENABLED=true` (default) and `gcc` present |
| `muslc` | Static builds |
| `static_wasm` | Static Darwin builds (goreleaser) |
| `cleveldb` | `GAIA_BUILD_OPTIONS=cleveldb` |

### Linker flags (injected at build time)

```
version.Name     = gaia
version.AppName  = gaiad
version.Version  = <git tag or branch-commit>
version.Commit   = <full git commit hash>
version.BuildTags
TMCoreSemVer     = <cometbft version>
```

---

## 12. Testing Strategy

### Test layers

| Layer | Location | Command | Notes |
|-------|----------|---------|-------|
| Unit | All `*_test.go` in non-e2e packages | `make test-unit` | Fast, no Docker |
| Unit + coverage | Same | `make test-unit-cover` | Generates `coverage.txt` |
| Race detection | Same | `make test-race` | Slower, catches data races |
| Simulation | `app/sim_test.go` | `make test-sim-*` (sims.mk) | Full state-machine simulation |
| Fuzz | `app/genesis_account_fuzz_test.go` | `go test -fuzz` | Genesis account fuzzing |
| Integration | `tests/integration/` | `go test ./tests/integration/...` | In-process multi-module tests |
| E2E | `tests/e2e/` | `make test-e2e` | Requires Docker; uses Hermes relayer |
| Interchain | `tests/interchain/` | Separate `go.mod`; GitHub Actions | Multi-chain topology tests |

### Mock generation

Mocks are generated with `mockery` (configured in `.mockery.yaml`):
```bash
make gen-mocks   # runs: go run github.com/vektra/mockery/v2 then gofmt
```

Generated mock files live in `x/liquid/types/mocks/`.

### Simulation

`sims.mk` defines targets for running randomized simulations:
- `make test-sim-full` — full simulation (used in CI on releases)
- `make test-sim-benchmark` — benchmarks
- Simulations use `go test -run TestFullAppSimulation` with seeds.

---

## 13. CI/CD Workflows

All workflows live in `.github/workflows/`.

| Workflow file | Trigger | Purpose |
|---------------|---------|---------|
| `test.yml` | PR, push to `main` | Unit + E2E tests |
| `lint.yml` | PR, push to `main` | golangci-lint |
| `interchain-test.yml` | PR, push | Multi-chain interchain tests |
| `nightly-tests.yml` | Nightly cron | Extended test suite |
| `sims.yml` | PR (sim label) | Simulation tests |
| `release-sims.yml` | Release tags | Full simulation on release |
| `release.yml` | Git tag `v*` | goreleaser → GitHub release |
| `docker-push.yml` | Push to `main` | Build + push Docker image |
| `deploy-docs.yml` | Push to `main` | Deploy Docusaurus docs |
| `dependabot-changelog.yml` | Dependabot PRs | Auto-add CHANGELOG entry |
| `codeql-analysis.yml` | Weekly | GitHub CodeQL security scan |
| `md-link-checker.yml` | PR | Check Markdown links |
| `stale.yml` | Daily | Mark/close stale issues/PRs |

### CI specifics

- Runners: `depot-ubuntu-22.04-4` (4-CPU Depot runners).
- Go version pinned to `1.25.7` in all workflows.
- Test jobs skip if no relevant files changed (`.go`, `go.mod`, `go.sum`, `Makefile`).
- Coverage artifacts uploaded per commit SHA.

### Mergify automation (`.mergify.yml`)

- **Automerge:** PRs with `A:automerge` label + ≥1 approval → squash merge to `main`.
- **Backport to `release/v27.x`:** Label `A:backport/v27.x`.
- **Backport to `release/v26.x`:** Label `A:backport/v26.x`.

---

## 14. Release Process

Gaia uses **semantic versioning** with these conventions:

| Version component | Meaning |
|------------------|---------|
| Major (X) | State-machine breaking change — requires coordinated network upgrade |
| Minor (Y) | Emergency release OR API breaking change (events, queries, CLI) |
| Patch (Z) | Bug fixes, non-breaking dependency bumps |

### Major release steps

1. Ensure `main` is feature-complete.
2. Create release branch: `git checkout -b release/vY.x`.
3. Add Mergify backport rule for `release/vY.x` in `.mergify.yml`.
4. Add version section to `CHANGELOG.md` and `RELEASE_NOTES.md`.
5. Audit, add tests, fix bugs — all fixes go to `main` first, then backport.
6. Cut release candidate: `make create-release TAG=vY.0.0-rc1`.
7. Test on public testnets.
8. If bugs found, fix on `main`, backport, cut new RC.
9. Final release **must have the same commit hash** as the last RC.
10. `make create-release TAG=vY.0.0` → triggers `release.yml` → goreleaser builds binaries for all platforms.

### Goreleaser targets

Produces static binaries for:
- `darwin/arm64` and `darwin/amd64` — linked against `libwasmvmstatic_darwin.a`
- `linux/amd64` and `linux/arm64` — linked against `libwasmvm_muslc.a` (fully static)

All binaries embed: `static_wasm`, `netgo`, `ledger` build tags.

### Non-major release steps

- Patch PRs go directly to `main` with backport labels.
- Mergify auto-backports to the appropriate `release/vY.x` branch.
- Tag and release from the release branch.

---

## 15. Upgrade Process (Validator/Node)

### Governance-gated upgrade

A software upgrade proposal passes on-chain, setting an upgrade plan at a future block height.

**Manual binary upgrade:**
1. Build or download the new binary.
2. Wait for node to halt at upgrade height (log: `ERR UPGRADE "vX.Y.Z" NEEDED at height: ...`).
3. Replace the binary; restart the node service.

**Cosmovisor upgrade (recommended):**
```
~/.gaia/cosmovisor/
├── current → upgrades/vX
├── genesis/bin/gaiad       # running version
└── upgrades/
    └── vX/bin/gaiad        # new binary placed here in advance
```

Sample Cosmovisor service env vars:
```
DAEMON_NAME=gaiad
DAEMON_HOME=/gaia/.gaia
DAEMON_ALLOW_DOWNLOAD_BINARIES=false  # always false for mainnet security
DAEMON_RESTART_AFTER_UPGRADE=true
UNSAFE_SKIP_BACKUP=true               # set false to auto-backup state
```

### Non-governance-gated upgrade

Binary replacement at any time — no chain halt needed. Used for minor/patch releases that are state-compatible.

---

## 16. State Compatibility Rules

Documented in `STATE-COMPATIBILITY.md`. Critical for patch and minor releases.

### In-scope (must be identical across patch/minor versions)

- Every msg's `ValidateBasic()` return value and error type
- Every msg's `MsgServer` method: gas consumed, error returned, state written
- All `AnteHandler` behavior in `DeliverTx` mode
- All `BeginBlock` / `EndBlock` logic

### Out-of-scope (can change freely)

- Event attributes
- Non-whitelisted queries
- CLI interfaces

### Validation mechanism

CometBFT hashes `AppHash` and `LastResultsHash` after every block. Any state-machine divergence causes consensus failure.

### Common sources of state-incompatibility (avoid in patch PRs)

1. Writing new state unconditionally in upgrade logic.
2. Changing protobuf field definitions in a non-additive way.
3. Returning a different error (or success) for the same input.
4. Non-deterministic gas consumption.
5. Using system time, randomness, or parallelism in message handlers.

---

## 17. Version Bump Procedure

Use the automated Makefile target when incrementing the major version:

```bash
make bump-version OLD_VERSION=v27.0.0 NEW_VERSION=v28.0.0
```

This updates:
- `go.mod` module path (`github.com/cosmos/gaia/vOLD` → `github.com/cosmos/gaia/vNEW`)
- All Go import paths across the codebase
- `tests/interchain/go.mod` import paths
- `.github/workflows/test.yml` (old binary download URL for upgrade tests)

After running, manually:
1. Create `app/upgrades/vNEW_0_0/` with `constants.go` and `upgrades.go`.
2. Add the upgrade to `Upgrades` in `app/app.go`.
3. Update `CHANGELOG.md` with the new version section.
4. Run `go mod tidy`.

---

## 18. Contributing Workflow

### Branch strategy

- All development targets `main`.
- Large features use long-lived feature branches, each with a tracking issue (EPIC).
- Release branches (`release/vY.x`) receive only backports — **never direct PRs** except in exceptional circumstances.
- PRs from Mergify backports are the only exception to the direct-to-main rule.

### PR requirements

1. **One thing per PR** — clear title, description, linked issue.
2. **Manageable size** — large contributions must be split into a series of small PRs.
3. **Changelog entry** — any production code change needs an entry in `CHANGELOG.md` under `UNRELEASED`.
4. **Tests** — new functionality must include tests at the appropriate layer.

### PR templates

Located in `.github/PULL_REQUEST_TEMPLATE/`:
- `production.md` — code changes
- `docs.md` — documentation only
- `others.md` — scripts, tooling, CI

### Changelog format

Sections: `FEATURES`, `BUG-FIXES`, `DEPENDENCIES`.
```markdown
## UNRELEASED

### FEATURES
- Description of feature ([#PRNUM](link))

### BUG-FIXES
- Fix description ([#PRNUM](link))

### DEPENDENCIES
- Bump [dep](link) from vOLD to vNEW ([#PRNUM](link))
```

### Code style

Enforced by golangci-lint v2.6.2 with these linters enabled:
`dogsled`, `errcheck`, `goconst`, `gocritic`, `gosec`, `govet`, `ineffassign`, `misspell`, `nakedret`, `nolintlint`, `revive`, `staticcheck`, `thelper`, `unconvert`, `unparam`, `unused`.

Formatting: `gofumpt` (stricter than `gofmt`).

```bash
make lint        # check
make lint-fix    # auto-fix where possible
make format      # gofumpt + lint fix
```

---

## 19. Local Development Quick Reference

### Single-node local chain (quickest)

```bash
make build
make start-localnet-ci
```

This initializes a single validator node with:
- Chain ID: `liveness`
- Home: `~/.gaiad-liveness`
- Min gas price: `0uatom`
- Keyring: `test` backend

### 4-node Docker localnet

```bash
make localnet-start   # builds image, initializes 4 nodes, starts docker-compose
make localnet-stop    # docker compose down
```

### Running tests

```bash
# Unit tests only
make test-unit

# With coverage
make test-unit-cover
cat coverage.txt | go tool cover -html=coverage.txt

# Race detection
make test-race

# E2E (requires Docker)
make docker-build-debug
make docker-build-hermes
make test-e2e
```

### Querying a running node

```bash
# gRPC
grpcurl -plaintext localhost:9090 list

# REST
curl http://localhost:1317/cosmos/bank/v1beta1/balances/<address>

# CLI
gaiad query bank balances <address> --node tcp://localhost:26657
```

### Useful gaiad commands

```bash
# Node info
gaiad status
gaiad version --long

# Keys
gaiad keys add mykey --keyring-backend test
gaiad keys list --keyring-backend test

# Governance
gaiad query gov proposals
gaiad tx gov submit-proposal <proposal.json> --from mykey

# IBC
gaiad query ibc channel channels
gaiad query ibc-transfer denom-trace <hash>

# Liquid staking
gaiad query liquid params
gaiad query liquid total-liquid-staked
gaiad tx liquid tokenize-share <validator> <amount> <reward-owner>

# ICS Provider
gaiad query provider consumer-chains
```

### Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `LEDGER_ENABLED` | `true` | Enable hardware Ledger signing support |
| `GAIA_BUILD_OPTIONS` | — | `cleveldb`, `nostrip`, etc. |
| `LINK_STATICALLY` | — | Set `true` for fully static build |
| `SKIP_MOD_VERIFY` | — | Skip `go mod verify` in `go.sum` target |
| `BUILDDIR` | `./build` | Output directory for `make build` |
| `GITHUB_TOKEN` | — | Required for `make ci-release` |

---

*Last updated: 2026-05-08 — tracks Gaia v28.0.0 (module path `github.com/cosmos/gaia/v28`)*
