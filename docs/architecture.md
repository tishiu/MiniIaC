# MiniIaC Architecture

This document reflects the current implementation in `cmd/` and `pkg/`.

## C4 Level 1: Context

```mermaid
flowchart LR
    U[Platform Engineer]

    subgraph S[MiniIaC System]
        M[miniac CLI]
    end

    FS[(Local Filesystem)]
    ST[(.goiac state store)]
    DD[(Docker Daemon)]
    YC[(YAML config files)]

    U -->|plan/apply/destroy/state| M
    M -->|read config| YC
    M -->|create/update/delete local_file| FS
    M -->|create/update/delete containers and networks| DD
    M -->|load/save state, lock, checksum| ST
```

## C4 Level 2: Containers

```mermaid
flowchart TB
    subgraph miniac[MiniIaC Process]
        MAIN[cmd/main.go\nCommand Router]
        CLI[pkg/cli\nUse-case handlers]
        PARSER[pkg/config\nYAML Parser]
        REC[pkg/reconciler\nPrepare + Commit]
        GRAPH[pkg/graph\nDAG + topo sort]
        REF[pkg/reference\nDependency extraction + interpolation]
        CAT[pkg/provider/catalog\nSchema + dispatch]
        STM[pkg/state\nLock + migrate + transact + save]
        PROV[pkg/provider/*\nlocal + docker providers]
        DRT[pkg/provider/docker/runtime\nDocker SDK adapter]
    end

    USER[User]
    YAML[(main.yaml)]
    SFS[(Filesystem)]
    SDIR[(.goiac/state.json + checksum + lock)]
    DOCKER[(Docker Daemon)]

    USER --> MAIN
    MAIN --> CLI
    CLI --> PARSER
    PARSER --> YAML
    CLI --> REC
    REC --> GRAPH
    GRAPH --> REF
    REC --> CAT
    REC --> STM
    CAT --> PROV
    PROV --> DRT
    DRT --> DOCKER
    PROV --> SFS
    STM --> SDIR
```

## C4 Level 3: Reconciliation Components

```mermaid
flowchart TB
    subgraph sg_cli["pkg/cli commands"]
        PLAN["Plan command"]
        APPLY["Apply command"]
        DESTROY["Destroy command"]
    end

    subgraph sg_prepare["reconciler prepare phase"]
        PREP["Reconciler.Prepare"]
        PDES["prepareDesired"]
        PDIS["prepareDestroy"]
        DIFF["ComputeDiff"]
        PLOAD["State.Load"]
        GBUILD["Graph.Build"]
        GDAG["ValidateDAG"]
        GREF["ValidateReferences"]
        GAPP["TopologicalSortReverse for apply"]
        GDES["TopologicalSort for destroy"]
        PLANOBJ["PreparedPlan changes and order"]
    end

    subgraph sg_commit["reconciler commit phase"]
        COMMIT["PreparedPlan.Commit"]
        CAPPLY["commitApply"]
        CDESTROY["commitDestroy"]
        STX["State.Transact"]
        SLOCK["WithLock and lock file"]
        TXIDX["Txn index logical to provider"]
        INTP["InterpolateReferences"]
        PUT["tx.Index.Put"]
        DEL["tx.Index.DeleteByProviderID"]
        SSAVE["State.Save and checksum"]
    end

    subgraph sg_catalog["pkg/provider catalog boundary"]
        CPREP["Catalog.Prepare"]
        CEXEC["Catalog.Execute"]
        CLOOK["Schema checks and provider lookup"]
    end

    subgraph sg_provider["provider implementations"]
        LFILE["local.FileProvider"]
        DCONT["docker.ContainerProvider"]
        DNET["docker.NetworkProvider"]
        DRT["DockerRuntime adapter"]
    end

    PLAN --> PREP
    APPLY --> PREP
    DESTROY --> PREP

    PREP --> PDES
    PREP --> PDIS

    PDES --> CPREP
    PDES --> PLOAD
    PDES --> DIFF
    PDES --> GBUILD
    GBUILD --> GDAG
    GBUILD --> GREF
    GBUILD --> GAPP
    PDES --> PLANOBJ

    PDIS --> PLOAD
    PDIS --> GBUILD
    PDIS --> GDES
    PDIS --> PLANOBJ

    APPLY --> COMMIT
    DESTROY --> COMMIT
    COMMIT --> CAPPLY
    COMMIT --> CDESTROY

    CAPPLY --> STX
    CDESTROY --> STX
    STX --> SLOCK
    STX --> TXIDX
    STX --> SSAVE

    CAPPLY --> INTP
    INTP --> CEXEC
    CAPPLY --> CEXEC
    CDESTROY --> CEXEC
    CEXEC --> CLOOK
    CAPPLY -->|create or update| PUT
    CAPPLY -->|delete in diff| DEL
    CDESTROY --> DEL
    PUT --> TXIDX
    DEL --> TXIDX

    CLOOK --> LFILE
    CLOOK --> DCONT
    CLOOK --> DNET
    DCONT --> DRT
    DNET --> DRT
```

## Sequence: `apply`

```mermaid
sequenceDiagram
    participant U as User
    participant CLI as cli.Apply
    participant P as config.Parser
    participant R as reconciler.Prepare
    participant C as provider.Catalog
    participant G as graph.Graph
    participant S as state.Manager
    participant T as state.Txn
    participant PV as Providers
    participant RT as DockerRuntime
    participant EXT as External Systems

    U->>CLI: miniac apply CONFIG_PATH
    CLI->>P: Parse(configPath)
    P-->>CLI: desired resources

    CLI->>R: Prepare(ModeApply, desired)
    loop each desired resource
        R->>C: Prepare(resource)
        C-->>R: validated/coerced resource
    end
    R->>S: Load()
    S-->>R: current state
    R->>R: ComputeDiff(desired, current)
    R->>G: Build(desired)
    G->>G: ValidateDAG + ValidateReferences
    G-->>R: TopologicalSortReverse()
    R-->>CLI: PreparedPlan(changes, order)

    CLI->>U: Print plan + approval prompt
    U-->>CLI: yes / --auto-approve
    CLI->>R: plan.Commit(ctx)
    R->>S: Transact(fn)
    S->>S: WithLock + Load
    S-->>T: tx(state,index)

    loop each resource in dependency-first order
        R->>R: InterpolateReferences(resource, tx.State)
        R->>C: Execute(Create/Update)
        C->>PV: Provider action
        alt docker_container or docker_network
            PV->>RT: Create/Update/Delete
            RT->>EXT: Docker API calls
        else local_file
            PV->>EXT: os.WriteFile / os.Remove
        end
        PV-->>C: ResourceState
        C-->>R: ResourceState
        R->>T: tx.Index.Put(logicalID, newState)
    end

    opt delete-only changes
        R->>C: Execute(Delete)
        C->>PV: Delete(resourceID)
        PV-->>C: ok
    end

    S->>S: Save(state.json + checksum)
    S-->>CLI: success
    CLI-->>U: Infrastructure updated successfully
```

## Sequence: `destroy`

```mermaid
sequenceDiagram
    participant U as User
    participant CLI as cli.Destroy
    participant R as reconciler.Prepare
    participant G as graph.Graph
    participant S as state.Manager
    participant C as provider.Catalog
    participant PV as Providers

    U->>CLI: miniac destroy
    CLI->>R: Prepare(ModeDestroy)
    R->>S: Load()
    S-->>R: current state
    R->>G: Build(resources from state)
    G-->>R: TopologicalSort() (dependents first)
    R-->>CLI: delete plan

    CLI->>U: Print resources + approval prompt
    U-->>CLI: yes / --auto-approve

    CLI->>R: plan.Commit()
    R->>S: Transact(fn)
    loop each delete change in order
        R->>C: Execute(Delete)
        C->>PV: Delete(providerID)
        PV-->>C: ok
    end
    S->>S: Save(updated state + checksum)
    CLI-->>U: All resources destroyed successfully
```

## State Transaction and Locking Flow

```mermaid
flowchart TD
    A[Start Transact] --> B{Context cancelled?}
    B -->|yes| X[Return error]
    B -->|no| C[Acquire .goiac/state.lock]
    C --> D{Lock exists?}
    D -->|stale/invalid| E[Remove lock and retry]
    E --> C
    D -->|active lock| X
    D -->|acquired| F[Load state.json]
    F --> G{Checksum file exists?}
    G -->|yes| H[Verify SHA-256]
    G -->|no| I[Migrate state version]
    H --> I
    I --> J[Build Txn index]
    J --> K[Run mutation function]
    K --> L{Mutation error?}
    L -->|yes| M[Save partial state by design]
    L -->|no| N[Save final state]
    M --> O[Write checksum]
    N --> O
    O --> P[Release lock]
    P --> Q[End]
```

## Core Diagram: Command Dispatch and CLI Routing

```mermaid
flowchart TD
    S["Program start"] --> PF["parseGlobalFlags"]
    PF --> A{"Remaining args empty"}
    A -->|yes| IM["runInteractiveMode"]
    A -->|no| NC["cli.NewCLI"]

    IM --> SC["Read command line"]
    SC --> EX{"exit or quit"}
    EX -->|yes| END["Leave interactive mode"]
    EX -->|no| EC["executeCommand"]

    NC --> EC

    EC --> CMD{"Command switch"}
    CMD -->|init| CINIT["CLI.Init"]
    CMD -->|plan| CPLAN["CLI.Plan"]
    CMD -->|apply| CAPP["CLI.Apply"]
    CMD -->|destroy| CDEST["CLI.Destroy"]
    CMD -->|state show| CSTATE["CLI.StateShow"]
    CMD -->|help| CHELP["printUsage"]
```

## Core Diagram: Diff Engine Decision Logic

```mermaid
flowchart TB
    START["Start ComputeDiff"] --> PASS1["Pass 1 over desired resources"]

    PASS1 --> EXISTS{"Resource exists in current state"}
    EXISTS -->|no| CREATE["Append ChangeTypeCreate"]
    EXISTS -->|yes| DIFF{"propertiesDiffer is true"}
    DIFF -->|yes| UPDATE["Append ChangeTypeUpdate"]
    DIFF -->|no| NOOP["Append ChangeTypeNoop"]
    CREATE --> PROC["Mark resource as processed"]
    UPDATE --> PROC
    NOOP --> PROC

    PROC --> PASS2["Pass 2 over current state"]
    PASS2 --> SEEN{"ID present in processed map"}
    SEEN -->|no| DELETE["Append ChangeTypeDelete"]
    SEEN -->|yes| KEEP["Keep resource"]
    DELETE --> DONE["Return changes list"]
    KEEP --> DONE
```

## Core Diagram: Reference Engine Resolution

```mermaid
sequenceDiagram
    participant RC as Resource
    participant EN as reference.Engine
    participant EX as ExtractDependencies
    participant RS as ResolveProperties
    participant ST as state.State

    RC->>EN: Process(resource, currentState)
    EN->>EX: Dependencies(resource.Properties)
    EX-->>EN: dependency IDs
    EN->>RS: Resolve(resource.Properties, currentState)

    loop each string interpolation token
        RS->>ST: lookup resource and attribute
        alt value found in state
            ST-->>RS: attribute value
        else missing value
            RS-->>RS: keep original token
        end
    end

    RS-->>EN: resolved properties
    EN-->>RC: Result dependencies and resolved resource
```

## Core Diagram: Dependency Ordering Semantics

```mermaid
flowchart TB
    WEB["docker_container web"] --> NET["docker_network app_net"]
    NOTE["Edge means dependent points to dependency"]
    NOTE --> WEB

    WEB --> A1["Apply mode uses TopologicalSortReverse"]
    A1 --> A2["Apply order app_net then web"]

    WEB --> D1["Destroy mode uses TopologicalSort"]
    D1 --> D2["Destroy order web then app_net"]
```

## Core Diagram: Provider Catalog Action State Machine

```mermaid
stateDiagram-v2
    [*] --> RequestReceived
    RequestReceived --> ValidateRequest

    ValidateRequest --> Rejected : invalid input
    ValidateRequest --> NoopReturn : action noop
    ValidateRequest --> CreateFlow : action create
    ValidateRequest --> UpdateFlow : action update
    ValidateRequest --> DeleteFlow : action delete

    CreateFlow --> ProviderCreate
    ProviderCreate --> Completed : returns ResourceState

    UpdateFlow --> ProviderUpdate
    ProviderUpdate --> Completed : returns ResourceState

    DeleteFlow --> ProviderDelete
    ProviderDelete --> Completed : returns nil state

    NoopReturn --> [*]
    Rejected --> [*]
    Completed --> [*]
```

## Core Diagram: State Data Model

```mermaid
erDiagram
    STATE_STORE ||--|| STATE : persists
    STATE ||--o{ RESOURCE_STATE : contains
    STATE_STORE ||--o| LOCK_INFO : lock_metadata

    STATE_STORE {
        string state_path
        string checksum_path
        string lock_path
    }

    STATE {
        int version
        string last_updated
    }

    RESOURCE_STATE {
        string logical_id_key
        string provider_id
        string resource_type
        json attributes
    }

    LOCK_INFO {
        datetime locked_at
        string locked_by
        int process_id
    }
```
