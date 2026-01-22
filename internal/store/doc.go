// Package store implements the data access layer for the assisted-migration-agent.
//
// This package provides persistent storage using DuckDB, combining locally-defined
// tables for agent configuration with tables created by duckdb_parser for VMware
// inventory data.
//
// # Architecture Overview
//
//	┌─────────────────────────────────────────────────────────────────┐
//	│                         Store (facade)                          │
//	├─────────────────────────────────────────────────────────────────┤
//	│       ConfigurationStore       │        InventoryStore          │
//	│              ▼                 │             ▼                  │
//	│        configuration           │         inventory              │
//	│           (local)              │           (local)              │
//	├────────────────────────────────┴────────────────────────────────┤
//	│                          VMStore                                │
//	│                             ▼                                   │
//	│    vinfo, vdisk, concerns (from duckdb_parser)                  │
//	└─────────────────────────────────────────────────────────────────┘
//
// # Data Sources
//
// Tables created by LOCAL MIGRATIONS (internal/store/migrations/sql/):
//
//	┌────────────────────┬─────────────────────────────────────────────┐
//	│  Table             │  Purpose                                    │
//	├────────────────────┼─────────────────────────────────────────────┤
//	│  configuration     │  Agent runtime config (agent_mode)          │
//	│  inventory         │  Raw inventory JSON blob with timestamps    │
//	│  schema_migrations │  Migration version tracking                 │
//	└────────────────────┴─────────────────────────────────────────────┘
//
// Tables created by DUCKDB_PARSER (parser.Init()):
//
//	┌─────────────────┬────────────────────────────────────────────────┐
//	│  Table          │  Purpose                                       │
//	├─────────────────┼────────────────────────────────────────────────┤
//	│  vinfo          │  Main VM info (ID, name, powerstate, cluster)  │
//	│  vcpu           │  CPU configuration per VM                      │
//	│  vmemory        │  Memory configuration per VM                   │
//	│  vdisk          │  Virtual disk info (capacity, RDM, sharing)    │
//	│  vnetwork       │  Network interfaces per VM                     │
//	│  vhost          │  ESXi host information                         │
//	│  vdatastore     │  Storage datastore information                 │
//	│  vhba           │  Host Bus Adapter information                  │
//	│  dvport         │  Distributed virtual port/VLAN info            │
//	│  concerns       │  Migration concerns/warnings per VM            │
//	└─────────────────┴────────────────────────────────────────────────┘
//
// # Initialization Flow
//
//	NewStore(db)
//	    ├── Creates duckdb_parser.Parser
//	    └── Initializes all sub-stores with QueryInterceptor
//
//	Store.Migrate(ctx)
//	    ├── parser.Init()     → Creates vinfo, vdisk, concerns, etc.
//	    └── migrations.Run()  → Creates configuration, inventory
//
// # Store Components
//
// # ConfigurationStore
//
// Persists agent runtime configuration in a single-row table.
//
// Schema:
//
//	configuration (
//	    id INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
//	    agent_mode VARCHAR DEFAULT 'disconnected'
//	)
//
// Methods:
//   - Get(ctx) → *models.Configuration
//   - Save(ctx, cfg) → error (uses UPSERT)
//
// # InventoryStore
//
// Stores raw inventory data as a JSON blob. Used by Console service to send
// inventory to console.redhat.com.
//
// Schema:
//
//	inventory (
//	    id INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
//	    data BLOB NOT NULL,
//	    created_at TIMESTAMP,
//	    updated_at TIMESTAMP
//	)
//
// Methods:
//   - Get(ctx) → *models.Inventory
//   - Save(ctx, data []byte) → error (uses UPSERT, updates updated_at)
//
// # VMStore
//
// Provides read access to VM inventory data. Uses a hybrid approach:
//   - List/Count: Direct SQL queries against duckdb_parser tables (vinfo, vdisk, concerns)
//   - Get: Uses parser.VMs() for full VM details with all relationships
//
// List Query Structure:
//
//	SELECT v."VM ID", v."VM", v."Powerstate", v."Cluster", v."Memory",
//	       COALESCE(d.total_disk, 0), COALESCE(c.issue_count, 0)
//	FROM vinfo v
//	LEFT JOIN (SELECT "VM ID", SUM("Capacity MiB") FROM vdisk GROUP BY "VM ID") d
//	LEFT JOIN (SELECT "VM_ID", COUNT(*) FROM concerns GROUP BY "VM_ID") c
//
// List Options:
//
// VMStore.List uses the functional options pattern. Each ListOption is a function
// that modifies the SQL query builder. Options can be combined for complex queries:
//
//	vms, err := store.VM().List(ctx,
//	    store.ByClusters("prod-cluster"),
//	    store.ByStatus("poweredOn"),
//	    store.ByIssues(1),
//	    store.WithSort([]store.SortParam{{Field: "name", Desc: false}}),
//	    store.WithLimit(50),
//	    store.WithOffset(0),
//	)
//
// Filtering Options:
//
//   - ByClusters(clusters ...string)
//     Filters VMs by cluster name. Multiple clusters use OR logic.
//     SQL: WHERE v."Cluster" IN (...)
//
//   - ByStatus(statuses ...string)
//     Filters VMs by power state. Multiple statuses use OR logic.
//     Values: "poweredOn", "poweredOff", "suspended"
//     SQL: WHERE v."Powerstate" IN (...)
//
//   - ByIssues(minIssues int)
//     Filters VMs with at least N migration concerns/issues.
//     SQL: WHERE COALESCE(c.issue_count, 0) >= minIssues
//
//   - ByDiskSizeRange(min, max int64)
//     Filters VMs by total disk capacity in MB. Range is [min, max).
//     SQL: WHERE total_disk >= min AND total_disk < max
//
//   - ByMemorySizeRange(min, max int64)
//     Filters VMs by memory in MB. Range is [min, max).
//     SQL: WHERE v."Memory" >= min AND v."Memory" < max
//
// Pagination Options:
//
//   - WithLimit(limit uint64)
//     Limits the number of results returned.
//     SQL: LIMIT limit
//
//   - WithOffset(offset uint64)
//     Skips the first N results (for pagination).
//     SQL: OFFSET offset
//
// Sorting Options:
//
//   - WithSort(sorts []SortParam)
//     Applies multi-field sorting. Each SortParam has Field and Desc (direction).
//     Always appends VM ID as tie-breaker for stable sorting.
//
//   - WithDefaultSort()
//     Sorts by VM ID ascending. Applied when no explicit sort is provided.
//
// Sort Field Mapping:
//
//	┌──────────────┬─────────────────────────────┐
//	│  API Field   │  Database Column            │
//	├──────────────┼─────────────────────────────┤
//	│  name        │  v."VM"                     │
//	│  vCenterState│  v."Powerstate"             │
//	│  cluster     │  v."Cluster"                │
//	│  diskSize    │  COALESCE(d.total_disk, 0)  │
//	│  memory      │  v."Memory"                 │
//	│  issues      │  issue_count                │
//	└──────────────┴─────────────────────────────┘
//
// # QueryInterceptor
//
// All database operations are wrapped with a QueryInterceptor that provides
// debug logging for all queries. This enables visibility into SQL execution
// without modifying individual store implementations.
//
// Logged operations:
//   - QueryRowContext
//   - QueryContext
//   - ExecContext
//
// # Design Patterns
//
// Single-Row Tables:
//   - Configuration and Inventory use CHECK (id = 1) constraint
//   - Guarantees only one record per logical entity
//   - Uses UPSERT pattern: INSERT ... ON CONFLICT (id) DO UPDATE
//
// Functional Options:
//   - VMStore uses ListOption functions for composable query building
//   - Each option modifies a squirrel.SelectBuilder
//   - Options can be combined for complex queries
//
// Separation of Concerns:
//   - Local tables: Agent state (configuration, raw inventory)
//   - Parser tables: Structured VMware inventory (VMs, hosts, datastores)
//   - VMStore bridges both: queries parser tables, returns domain models
package store
