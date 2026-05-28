# Arso Architecture

**Project:** Arso  
**Full name:** Autonomous Remote Space Observatory  
**Repository:** `openarso/arso`  
**Document type:** Architecture specification  
**Status:** Draft

---

## 1. Purpose

This document describes the technical architecture of Arso.

Arso is an open-source autonomous remote space and weather observatory platform. Its architecture must support a progressive path:

```txt
predict → measure → capture → store → record → survive → sync → receive → analyze → share
```

The architecture must stay local-first. Arso should become a reliable single-node observatory before it becomes a distributed network.

---

## 2. Architectural Principles

## 2.1 Local First

An Arso node must be useful without cloud services, ARSO-Net, public registry, or external synchronization.

A local node should be able to:

- predict ISS and satellite passes;
- monitor weather;
- capture all-sky images;
- store data locally;
- recover from restart;
- expose local CLI/API/dashboard access.

## 2.2 Hardware-Light First Feature

The first real feature should be satellite and ISS pass prediction.

This requires only:

- a computer or Raspberry Pi;
- observer location;
- time/timezone handling;
- orbital data source;
- CLI and configuration.

This gives the project an early space-oriented feature without waiting for cameras, sensors, SDR, or telescope hardware.

## 2.3 Simulated Mode Is Mandatory

Every major hardware-facing component should have a simulated provider.

Simulated mode allows contributors to work without:

- camera;
- weather station;
- telescope;
- SDR receiver;
- motorized mount.

Simulated mode is not a toy. It is a development and testing requirement.

## 2.4 Stable Data Before Distributed Features

Arso must define durable local data models before adding synchronization, computer vision, or ARSO-Net.

Distributed systems amplify bad local models. The project should avoid sharing poorly structured data.

## 2.5 Degraded Operation Before Sync

Synchronization only matters if the node can survive ordinary failures.

Before sync, Arso should handle:

- network loss;
- process restart;
- interrupted captures;
- partial writes;
- storage verification;
- basic recovery states.

## 2.6 AI Later, Dataset First

Computer vision should come after Arso produces useful captures and event metadata.

Early AI would mostly be demo logic. The system should first collect structured images, weather, passes, and events.

## 2.7 ARSO-Net Is Optional

ARSO-Net must never be required for local operation.

A node may eventually share selected metadata, but local observation remains independent.

---

## 3. High-Level System View

At a high level, Arso consists of:

```txt
+-------------------+
|       CLI         |
+-------------------+
          |
          v
+-------------------+        +-------------------+
|     API Server    | <----> |    Web Dashboard  |
+-------------------+        +-------------------+
          |
          v
+-------------------+
|    Edge Agent     |
+-------------------+
          |
          v
+--------------------------------+
| Providers / Hardware Adapters  |
+--------------------------------+
```

The core local node uses:

```txt
Satellite Prediction
Weather Monitoring
All-Sky Capture
Local Storage
Event Recording
Degraded Operation
```

Later optional modules add:

```txt
Telescope Capture
Private Synchronization
SDR Reception
Computer Vision
ARSO-Net Sharing
Community Registry
Distributed Campaigns
```

---

## 4. Runtime Modes

Arso should support several runtime modes.

## 4.1 Simulated Local Mode

For development and testing without hardware.

```txt
CLI → Local API/Agent → Simulated Providers → Local Storage
```

Capabilities:

- simulated node status;
- simulated weather;
- simulated all-sky capture;
- real satellite prediction;
- local file/database storage.

This should be the default contributor mode.

## 4.2 Local Hardware Mode

For a real node running on a Raspberry Pi, mini-PC, or Linux machine.

```txt
CLI/Dashboard → API → Edge Agent → Camera / Sensors / Storage
```

Capabilities:

- real or simulated satellite prediction;
- real or API-based weather;
- real all-sky camera capture;
- local metadata and image storage;
- basic dashboard.

## 4.3 Remote Private Mode

For a user operating their own node remotely.

```txt
Remote CLI/Dashboard → Secured API → Edge Agent → Local Hardware
```

Capabilities:

- authenticated remote status;
- remote logs;
- remote capture trigger;
- safe remote commands;
- private sync target.

This is not ARSO-Net.

## 4.4 ARSO-Net Mode

For optional public/community metadata sharing.

```txt
Local Node → Sharing Policy → ARSO-Net Gateway/Registry
```

Capabilities:

- approximate node profile sharing;
- selected metadata export;
- public/private visibility;
- distributed observation campaigns.

This comes later and must be opt-in.

---

## 5. Main Components

## 5.1 CLI

The CLI is the primary operator and developer interface.

### Responsibilities

The CLI should:

- initialize configuration;
- manage observer location;
- query node status;
- predict satellite/ISS passes;
- query weather;
- trigger captures;
- compose pipeline workflows;
- produce structured JSON/JSONL output;
- consume structured input from stdin;
- inspect local data;
- run diagnostics;
- manage sync later;
- interact with ARSO-Net later.

### Early Commands

```bash
arso version
arso config init
arso config show
arso node status
arso find ISS
arso pass next --target ISS
arso pass list --target ISS --next 24h
arso follow ISS --dry-run
arso weather current
arso plan capture allsky --from -
arso capture allsky
arso run -
arso data stats
```

### Design Rules

The CLI must not contain hardware-specific business logic.

It should call:

- local libraries for pure calculations where appropriate;
- the local API or agent for node operations;
- stable service interfaces for actions.

The CLI should support:

```bash
--output text
--output json
--output jsonl
--from -
```

Text is for humans. JSON and JSONL are for automation. `--from -` is for pipeline input from stdin.

---

## 5.1.1 CLI Pipeline Architecture

Arso's CLI should be designed as a composable pipeline system inspired by Unix tools.

The goal is to support workflows such as:

```bash
arso pass next ISS -o json | arso plan capture allsky --from - | arso run -
```

This is the scalable version of the human idea:

```txt
locate ISS | take picture
```

The CLI should make simple commands easy while still allowing advanced users to compose complex observation workflows.

---

### Pipeline Design Philosophy

Each command should do one clear job.

```txt
query → plan → validate/filter → execute → store/report
```

Commands should be composable through standard streams:

```txt
stdout = useful command output
stderr = logs, warnings, and errors
stdin  = optional structured input
```

This means logs must not pollute `stdout` when structured output is requested.

---

### Structured Output

Commands that produce machine-readable data must support:

```bash
-o json
--output json
```

For streams or lists, commands should support:

```bash
-o jsonl
--output jsonl
```

Use JSON for one object.

Use JSONL for many objects.

Example single-object output:

```bash
arso pass next ISS -o json
```

```json
{
  "type": "satellite_pass",
  "target": "ISS",
  "observer": {
    "latitude": 48.8566,
    "longitude": 2.3522,
    "timezone": "Europe/Paris"
  },
  "startTime": "2026-05-28T20:14:32Z",
  "peakTime": "2026-05-28T20:17:12Z",
  "endTime": "2026-05-28T20:20:14Z",
  "maxElevationDeg": 64,
  "visible": true
}
```

Example stream output:

```bash
arso pass list ISS --tonight -o jsonl
```

```jsonl
{"type":"satellite_pass","target":"ISS","startTime":"2026-05-28T20:14:32Z","maxElevationDeg":64}
{"type":"satellite_pass","target":"ISS","startTime":"2026-05-29T01:02:11Z","maxElevationDeg":31}
```

---

### Structured Input

Commands that consume structured data must support reading from stdin using:

```bash
--from -
```

Example:

```bash
arso pass next ISS -o json | arso plan capture allsky --from -
```

`--from -` means:

```txt
read structured input from stdin
```

Commands may also support reading from a file:

```bash
arso plan capture allsky --from pass.json
```

This makes workflows reproducible and scriptable.

---

### Piped Object Contract

Every structured object emitted by the CLI should include a `type` field.

Examples:

```json
{
  "type": "satellite_pass"
}
```

```json
{
  "type": "weather_measurement"
}
```

```json
{
  "type": "capture_plan"
}
```

```json
{
  "type": "capture_result"
}
```

The `type` field allows downstream commands to validate input and fail clearly.

For example:

```bash
arso capture allsky --from -
```

may accept:

```txt
satellite_pass
capture_plan
observation_plan
```

but reject:

```txt
weather_measurement
node_status
analysis_result
```

with a clear error:

```txt
error: expected satellite_pass, capture_plan, or observation_plan from stdin, got weather_measurement
```

---

### Command Categories

Arso CLI commands should be grouped by pipeline role.

#### Query Commands

Query commands produce facts or state.

Examples:

```bash
arso node status
arso pass next ISS
arso pass list ISS --next 24h
arso weather current
arso weather safe-to-observe
arso captures list
```

They should not mutate hardware or local state except for harmless cache updates.

#### Planning Commands

Planning commands convert facts into executable plans.

Examples:

```bash
arso plan capture allsky --from -
arso plan observe --from -
arso plan follow --from -
```

Example:

```bash
arso pass next ISS -o json | arso plan capture allsky --from -
```

Output:

```json
{
  "type": "capture_plan",
  "target": "ISS",
  "captureType": "allsky",
  "triggerTime": "2026-05-28T20:14:32Z",
  "durationSeconds": 342,
  "reason": "ISS visible pass"
}
```

#### Validation and Filter Commands

Validation/filter commands accept structured input and either pass it through, transform it, or reject it.

Examples:

```bash
arso require weather-safe
arso filter elevation --min 30
arso filter visible
```

Example:

```bash
arso pass list ISS --tonight -o jsonl | arso filter elevation --min 30
```

These commands should preserve pipeline compatibility.

#### Execution Commands

Execution commands perform actions.

Examples:

```bash
arso run -
arso job create --from -
arso capture allsky
```

Execution commands should support:

```bash
--dry-run
```

when possible.

---

### Recommended Early Pipeline Workflows

#### Find the Next ISS Pass

```bash
arso pass next ISS
```

#### Get the Next ISS Pass as JSON

```bash
arso pass next ISS -o json
```

#### Plan an All-Sky Capture for the Next ISS Pass

```bash
arso pass next ISS -o json | arso plan capture allsky --from -
```

#### Execute the Planned Capture

```bash
arso pass next ISS -o json \
  | arso plan capture allsky --from - \
  | arso run -
```

#### Create a Scheduled Job Instead of Running Immediately

```bash
arso pass next ISS -o json \
  | arso plan capture allsky --from - \
  | arso job create --from -
```

#### Capture Next ISS Pass Only if Weather Is Safe

```bash
arso pass next ISS -o json \
  | arso require weather-safe \
  | arso plan capture allsky --from - \
  | arso job create --from -
```

#### Capture All Visible ISS Passes Tonight Above 30 Degrees

```bash
arso pass list ISS --tonight -o jsonl \
  | arso filter elevation --min 30 \
  | arso plan capture allsky --from - \
  | arso job create --from -
```

---

### Human-Friendly Shortcuts

Arso may provide shortcut commands for common workflows.

Example:

```bash
arso observe ISS --allsky
```

could internally represent:

```bash
arso pass next ISS -o json \
  | arso plan capture allsky --from - \
  | arso run -
```

Shortcuts are allowed, but they should not replace composable primitives.

The architecture should favor both:

```txt
simple commands for beginners
composable pipelines for advanced users
```

---

### Standard CLI Flags

The following flags should be standardized across commands where relevant:

```bash
-o, --output text|json|jsonl
--from -
--dry-run
--pretty
--quiet
--verbose
```

Recommended behavior:

| Flag | Meaning |
|---|---|
| `--output text` | Human-readable output |
| `--output json` | One structured JSON object |
| `--output jsonl` | Stream of structured JSON objects |
| `--from -` | Read structured input from stdin |
| `--dry-run` | Show what would happen without executing |
| `--pretty` | Pretty-print structured output |
| `--quiet` | Reduce non-essential output |
| `--verbose` | Increase diagnostic output on stderr |

---

### Pipeline Compatibility Rules

A command is pipeline-compatible when:

1. it does not mix logs with structured output;
2. it uses `stdout` for data;
3. it uses `stderr` for logs/errors;
4. it supports `--output json` when producing one object;
5. it supports `--output jsonl` when producing streams;
6. it supports `--from -` when consuming structured input;
7. it includes a `type` field in structured output;
8. it validates input object types;
9. it returns non-zero exit codes on failure;
10. it provides useful error messages.

---

### Exit Codes

Arso should define stable exit code categories.

Suggested initial model:

| Code | Meaning |
|---:|---|
| 0 | Success |
| 1 | General error |
| 2 | Invalid CLI usage |
| 3 | Invalid input object |
| 4 | Provider unavailable |
| 5 | Unsafe observation conditions |
| 6 | Hardware unavailable |
| 7 | Storage error |
| 8 | Network unavailable |
| 9 | Authentication/authorization error |

These do not need to be perfect from day one, but the project should avoid random exit behavior.

---

### Architectural Constraint

The pipeline system must not become a shell language or workflow engine too early.

Arso should use ordinary shell composition first.

This is good:

```bash
arso pass next ISS -o json | arso plan capture allsky --from - | arso run -
```

This is too early:

```txt
custom Arso workflow DSL
custom distributed workflow engine
complex visual automation builder
```

A real scheduler can come later, but the CLI pipeline should remain simple, inspectable, and scriptable.

---

## 5.2 API Server

The API server exposes local node data and actions to the CLI, dashboard, and future integrations.

### Responsibilities

The API should:

- expose health and node status;
- expose capabilities;
- expose pass prediction;
- expose weather;
- expose captures;
- expose local events;
- expose diagnostics;
- manage remote private access later.

### Early Endpoints

```http
GET    /health
GET    /api/v1/node/status
GET    /api/v1/node/capabilities

GET    /api/v1/satellites/{target}/passes
GET    /api/v1/satellites/{target}/passes/next

GET    /api/v1/weather/current
GET    /api/v1/weather/history

POST   /api/v1/captures/allsky
GET    /api/v1/captures
GET    /api/v1/captures/{captureId}

GET    /api/v1/events
GET    /api/v1/events/{eventId}

GET    /api/v1/storage/status
GET    /api/v1/diagnostics
```

### Design Rules

The API should:

- validate inputs;
- return structured errors;
- expose OpenAPI documentation;
- avoid public/multi-tenant assumptions in early versions;
- work locally without cloud dependencies.

---

## 5.3 Edge Agent

The edge agent represents a physical or simulated observatory node.

### Responsibilities

The agent should:

- report node health;
- report capabilities;
- manage hardware providers;
- execute local actions;
- perform captures;
- read weather providers;
- write local metadata;
- expose diagnostics;
- handle degraded operation.

### Agent Boundary

The agent is responsible for operational interaction with the node.

It should own:

- camera adapter calls;
- weather adapter calls;
- local hardware state;
- process-level diagnostics;
- capture execution;
- local recovery behavior.

It should not own:

- public registry logic;
- community features;
- heavy AI model training;
- SaaS-style account management.

### Capability Model

```json
{
  "nodeId": "local-node-001",
  "capabilities": {
    "satellitePrediction": true,
    "weather": true,
    "allSkyCamera": true,
    "telescopeCamera": false,
    "eventRecording": true,
    "degradedOperation": true,
    "sync": false,
    "sdr": false,
    "computerVision": false,
    "arsoNet": false
  }
}
```

---

## 5.4 Web Dashboard

The dashboard is a local-first visual interface.

### Responsibilities

The dashboard should display:

- node status;
- capabilities;
- next ISS/satellite passes;
- weather;
- recent all-sky captures;
- local events;
- storage status;
- diagnostics;
- later sync and SDR state.

### Design Rules

The dashboard should not become the core logic layer.

It should call the API and display state.

---

## 5.5 Satellite Prediction Module

The satellite prediction module is the first real feature module.

### Responsibilities

It should:

- use observer location;
- ingest or load orbital elements;
- predict visible passes;
- compute azimuth/elevation points;
- expose next-pass and pass-list results;
- support dry-run follow plans.

### Inputs

```txt
observer latitude
observer longitude
observer elevation
time range
target satellite
orbital data
visibility constraints
```

### Outputs

```txt
start time
peak time
end time
max elevation
direction
duration
azimuth/elevation time series
visibility flag
```

### Design Rules

Prediction must be separated from hardware movement.

This command:

```bash
arso follow ISS --dry-run
```

may produce a movement plan, but it must not require a motorized mount.

---

## 5.6 Weather Module

The weather module provides environmental information and observation safety hints.

### Providers

The module should support:

```txt
simulated provider
public weather API provider
local sensor provider
```

### Responsibilities

It should provide:

- current weather;
- weather history;
- safe-to-observe baseline;
- weather metadata for captures;
- later scheduling constraints.

### Safe-to-Observe Decision

The early safe-to-observe decision should be rule-based.

Example:

```json
{
  "safeToObserve": true,
  "reasons": [],
  "conditions": {
    "cloudCoverPercent": 12,
    "rainDetected": false,
    "windSpeedKph": 7.2
  }
}
```

The project should avoid claiming scientific or safety-grade precision too early.

---

## 5.7 Capture Module

The capture module manages image acquisition.

### First Target

The first real capture target is:

```txt
all-sky image capture
```

Telescope capture comes later.

### Responsibilities

The capture module should:

- trigger image capture;
- support simulated capture;
- support one real camera path first;
- write image files locally;
- generate capture metadata;
- attach weather metadata where available;
- attach satellite pass context where relevant.

### Capture Types

```txt
allsky
telescope
```

### Capture Metadata

```json
{
  "captureId": "cap_20260528_221432",
  "type": "allsky",
  "nodeId": "local-node-001",
  "capturedAt": "2026-05-28T22:14:32Z",
  "weatherMeasurementId": "wth_20260528_221430",
  "relatedPassId": "pass_iss_20260528_2214",
  "file": {
    "path": "data/captures/2026/05/28/cap_20260528_221432.jpg",
    "mimeType": "image/jpeg"
  }
}
```

---

## 5.8 Local Storage Module

The storage module is responsible for durable local persistence.

### Responsibilities

It should store:

- node metadata;
- observer location;
- satellite predictions;
- weather measurements;
- captures;
- capture files;
- events;
- diagnostics;
- jobs later;
- sync state later;
- SDR metadata later;
- analysis results later.

### Storage Layers

Arso should separate metadata and files.

```txt
Metadata → SQLite or PostgreSQL
Files    → local filesystem or S3-compatible storage
```

For early versions, a simple local setup is enough:

```txt
metadata: SQLite
files:    local filesystem
```

PostgreSQL and MinIO can be added for more production-like deployments.

### Local Data Layout

Example:

```txt
data/
├── arso.db
├── captures/
│   └── 2026/
│       └── 05/
│           └── 28/
│               └── cap_20260528_221432.jpg
├── events/
├── exports/
└── logs/
```

### Design Rules

Local storage must be the source of truth until explicit sync exists.

Failed sync must never corrupt local data.

---

## 5.9 Event Module

The event module records meaningful things that happen on the node.

### Event Types

```txt
satellite_pass_predicted
weather_measured
capture_created
capture_failed
network_unavailable
power_event
storage_warning
job_started
job_failed
sdr_recording_created
analysis_completed
```

### Responsibilities

It should:

- record structured events;
- link events to captures, weather, passes, and jobs;
- support later event review;
- prepare for future anomaly detection.

### Event Example

```json
{
  "eventId": "evt_20260528_221432",
  "type": "capture_created",
  "nodeId": "local-node-001",
  "timestamp": "2026-05-28T22:14:32Z",
  "relatedCaptureId": "cap_20260528_221432",
  "severity": "info"
}
```

---

## 5.10 Scheduler and Job Module

The scheduler comes after the basic local observation path works.

### Responsibilities

It should:

- create jobs;
- queue jobs;
- execute jobs through the agent;
- check weather constraints;
- use satellite pass windows;
- record job logs;
- support cancellation and interruption.

### Job States

```txt
created
queued
running
succeeded
failed
cancelled
interrupted
```

### Example Jobs

```txt
capture all-sky every 5 minutes
capture all-sky during next ISS pass
record weather every minute
capture telescope target manually
```

---

## 5.11 Degraded Operation Module

Degraded operation is a cross-cutting architecture concern.

### Failure Conditions

Arso should handle:

- no internet;
- API unavailable;
- weather provider unavailable;
- orbital data provider unavailable;
- camera unavailable;
- process restart;
- partial capture write;
- storage nearing full;
- power interruption, when detectable.

### Required Behaviors

The node should:

- continue local operation when internet is unavailable;
- cache needed orbital data where possible;
- mark incomplete captures clearly;
- verify local data integrity;
- keep local metadata consistent;
- expose recovery status;
- avoid destructive automatic recovery.

### Diagnostics Commands

```bash
arso diagnostics
arso data verify
arso node recovery-status
```

---

## 5.12 Synchronization Module

Synchronization is private and owner-controlled at first.

It is not ARSO-Net.

### Responsibilities

It should:

- sync metadata;
- sync selected files;
- resume after interruption;
- track sync state;
- avoid corrupting local data;
- allow user-defined sync targets.

### Sync Targets

```txt
local server
NAS
remote VPS
S3-compatible storage
another private Arso instance
```

### Design Rules

The local node remains the source of truth.

Raw images must not be shared publicly by default.

---

## 5.13 SDR Module

The SDR module adds radio reception.

### Responsibilities

It should:

- detect supported SDR devices;
- configure frequency;
- record selected signals;
- store signal metadata;
- optionally generate spectrum/waterfall artifacts;
- remain disabled by default.

### Important Boundary

Arso should receive radio signals only.

The project should not support radio transmission as an early feature.

---

## 5.14 Vision Module

The vision module analyzes accumulated images and events.

### Responsibilities

It may provide:

- image quality analysis;
- cloud/blur/overexposure detection;
- Moon/Sun/star-field classification;
- satellite or aircraft trail candidates;
- anomaly candidate scoring;
- manual review workflow.

### Design Rules

Vision results must be stored as metadata.

Vision should produce candidate/probability outputs, not unsupported scientific certainty.

The system must work without the vision module.

---

## 5.15 ARSO-Net Module

ARSO-Net is the future distributed network layer.

### Responsibilities

It may provide:

- metadata sharing;
- public/private node profiles;
- approximate location sharing;
- node capability discovery;
- distributed observation campaigns;
- dataset export.

### Privacy Rules

Not shared by default:

- exact private location;
- raw images;
- credentials;
- personal information;
- hardware control access;
- private logs.

### Design Rules

ARSO-Net must be opt-in.

Local Arso operation must not depend on ARSO-Net.

---

## 6. Suggested Monorepo Structure

```txt
arso/
├── apps/
│   ├── cli/
│   ├── api/
│   ├── web/
│   └── edge-agent/
│
├── services/
│   ├── tracker/
│   ├── weather/
│   ├── vision/
│   └── sdr/
│
├── packages/
│   ├── protocol/
│   ├── sdk/
│   └── common/
│
├── infra/
│   ├── compose/
│   └── docker/
│
├── docs/
│   ├── vision.md
│   ├── roadmap.md
│   ├── architecture.md
│   └── cahier-des-charges.md
│
├── firmware/
│   └── README.md
│
├── data/
│   └── .gitkeep
│
├── .github/
│   └── workflows/
│
├── Makefile
├── README.md
└── LICENSE
```

---

## 7. Internal Package Boundaries

## 7.1 `packages/protocol`

Contains shared schemas and contracts.

Examples:

```txt
NodeStatus
NodeCapability
ObserverLocation
SatellitePass
WeatherMeasurement
CaptureMetadata
Event
Job
SyncState
AnalysisResult
```

This package should avoid business logic.

## 7.2 `packages/sdk`

Client library used by CLI, dashboard, or external integrations.

Responsibilities:

- API client;
- typed request/response models;
- authentication support later;
- retry rules for safe read operations.

## 7.3 `packages/common`

Shared utilities that are genuinely common.

Examples:

- time helpers;
- logging helpers;
- configuration helpers;
- ID generation;
- filesystem helpers.

Do not turn `common` into a dumping ground.

---

## 8. Data Architecture

## 8.1 Core Entities

```txt
Node
ObserverLocation
Capability
SatelliteTarget
SatellitePass
WeatherMeasurement
Capture
CaptureFile
Event
Job
SyncState
SdrRecording
AnalysisResult
ArsoNetProfile
```

## 8.2 Entity Relationships

```txt
Node
 ├── ObserverLocation
 ├── Capabilities
 ├── WeatherMeasurements
 ├── SatellitePasses
 ├── Captures
 │    ├── CaptureFile
 │    ├── WeatherMeasurement
 │    └── SatellitePass
 ├── Events
 ├── Jobs
 ├── SdrRecordings
 └── AnalysisResults
```

## 8.3 Metadata vs Files

Metadata should be queryable.

Files should be stored outside the database.

```txt
Database:
- captures
- weather measurements
- satellite passes
- events
- jobs

Filesystem/Object Storage:
- images
- SDR recordings
- waterfall artifacts
- exports
- logs, if retained
```

---

## 9. Configuration Architecture

Configuration should be explicit and local.

Example:

```yaml
node:
  id: local-node-001
  name: Local Arso Node

observer:
  latitude: 48.8566
  longitude: 2.3522
  elevation_m: 35
  timezone: Europe/Paris
  public_location_precision: city

providers:
  satellite:
    source: celestrak
    cache_ttl_hours: 24
  weather:
    type: simulated
  camera:
    type: simulated
  storage:
    type: filesystem
    path: ./data

api:
  bind_address: 127.0.0.1
  port: 8080
```

## 9.1 Configuration Rules

- Secrets must not be committed.
- Exact location must not be public by default.
- Simulated providers should be easy to enable.
- Provider configuration should be validated at startup.
- CLI should provide `arso config init`.

---

## 10. Communication Architecture

## 10.1 Early Communication

Use simple communication first.

```txt
CLI → API/Agent: HTTP
Dashboard → API: HTTP
API → Agent: HTTP or local process call
Agent → Providers: direct adapter calls
```

## 10.2 Later Communication

Later versions may introduce:

```txt
MQTT for node events
gRPC for internal service communication
message queues for heavy async workloads
```

But these should not be default early dependencies.

## 10.3 Avoid Early

Avoid early dependency on:

- Kafka;
- Kubernetes service mesh;
- distributed consensus;
- complex event buses;
- cloud-only queues.

---

## 11. Deployment Architecture

## 11.1 Development Deployment

Early development should work with:

```txt
local binary execution
Makefile/task runner
Docker Compose
simulated providers
SQLite/local filesystem
```

## 11.2 Raspberry Pi Deployment

A Raspberry Pi node should eventually run:

```txt
arso-agent
optional arso-api
optional arso-web
local data directory
systemd service or container
```

## 11.3 Local Server Deployment

A more complete local deployment may run:

```txt
API server
edge agent
web dashboard
SQLite or PostgreSQL
filesystem or MinIO
```

## 11.4 Production-Like Private Deployment

Later private deployments may use:

```txt
PostgreSQL
MinIO/S3-compatible storage
reverse proxy
TLS
authenticated API
private sync target
```

Kubernetes and Helm may come later, but should not be required for v1.

---

## 12. Security Architecture

## 12.1 Early Security Rules

Even early versions should follow these rules:

- do not log secrets;
- keep exact location private by default;
- bind local services to localhost by default;
- require explicit configuration for remote access;
- validate API inputs;
- separate public metadata from private node data.

## 12.2 Authentication

Authentication becomes important for remote operation.

Possible early model:

```txt
local mode: no auth or local token
remote private mode: API token
ARSO-Net mode: separate sharing identity
```

## 12.3 Location Privacy

Arso needs exact location for prediction.

Arso must not publish exact location by default.

Public profile example:

```json
{
  "publicLocation": "Paris area, France",
  "precision": "city"
}
```

---

## 13. Observability Architecture

Arso should expose enough information to debug local failures.

## 13.1 Logs

Logs should be structured.

Important fields:

```txt
timestamp
level
component
nodeId
eventType
jobId
captureId
errorCode
message
```

## 13.2 Health Checks

Health checks should cover:

- process status;
- storage availability;
- provider status;
- camera availability;
- weather provider status;
- orbital data cache state;
- network state.

## 13.3 Diagnostics

Diagnostics should be available through:

```bash
arso diagnostics
```

and later:

```http
GET /api/v1/diagnostics
```

---

## 14. Failure and Recovery Architecture

## 14.1 Failure States

The system should represent degraded states explicitly.

Examples:

```txt
online
degraded
offline
recovering
maintenance
```

## 14.2 Capture Failure Handling

If a capture fails:

- record an event;
- preserve partial file only if useful;
- mark metadata as failed or interrupted;
- do not pretend capture succeeded;
- expose the failure reason.

## 14.3 Storage Failure Handling

If storage is unavailable or almost full:

- stop non-essential writes;
- emit warning events;
- expose diagnostics;
- avoid corrupting metadata;
- avoid destructive cleanup unless explicitly configured.

## 14.4 Network Failure Handling

If network is unavailable:

- keep local operation running;
- use cached satellite data if valid;
- mark external providers unavailable;
- delay sync;
- expose state clearly.

---

## 15. Extension Architecture

Arso should support new providers without rewriting core logic.

## 15.1 Provider Interfaces

Expected provider categories:

```txt
SatelliteDataProvider
WeatherProvider
CameraProvider
StorageProvider
SdrProvider
VisionProvider
SyncProvider
```

## 15.2 Provider Rules

Providers should:

- expose capabilities;
- return structured errors;
- support simulated implementation where relevant;
- avoid leaking vendor-specific details into core models.

## 15.3 Plugin System

A full plugin system is not required early.

For v0/v1, static provider registration is acceptable.

A dynamic plugin system can be considered later if community integrations grow.

---

## 16. Technology Choices

These choices are recommendations, not permanent constraints.

| Area | Recommended Start |
|---|---|
| CLI | Go |
| Edge agent | Go |
| API server | Go, Spring Boot, or FastAPI |
| Web dashboard | Angular |
| Satellite prediction | Dedicated tracker module/library |
| Weather providers | Simulated first, public API later, local sensors later |
| Metadata storage | SQLite first, PostgreSQL later |
| File storage | Local filesystem first, MinIO/S3 later |
| Vision service | Python later |
| SDR tooling | Native SDR tools/library wrapper later |
| Local deployment | Makefile + Docker Compose |
| CI | GitHub Actions |

## 16.1 Stack Warning

The stack should serve the project.

Do not add technologies to make the architecture look impressive. Add them when the operational need is real.

---

## 17. Architectural Milestones

## v0.1

Architecture documentation and monorepo skeleton.

## v0.2

CLI and local configuration.

## v0.3

Edge agent and simulated node.

## v0.4

Satellite and ISS pass prediction.

## v0.5

Weather monitoring.

## v0.6

All-sky capture.

## v0.7

Local storage and metadata.

## v0.8

Local API and dashboard.

## v0.9

Degraded operation.

## v1.x

Reliable single-node observatory.

## v2.x

Computer vision and ARSO-Net.

---

## 18. Explicit Non-Goals Before v2

Before v2, Arso should not prioritize:

- public ARSO-Net registry;
- public remote control;
- mesh networking;
- distributed consensus;
- AI-first product direction;
- custom model training;
- mandatory cloud services;
- Kubernetes-first deployment;
- Kafka-first event architecture;
- supporting every telescope/camera/SDR;
- scientific-grade claims without calibration.

---

## 19. Architecture Decision Rules

When making an architectural decision, prefer the option that:

1. helps `arso find ISS` become useful quickly;
2. keeps CLI commands composable through standard streams;
3. keeps local operation independent;
4. supports simulated mode;
5. preserves local data safely;
6. avoids premature distributed complexity;
7. makes hardware integration replaceable;
8. keeps privacy boundaries explicit;
9. can be explained to a new contributor.

If a decision does not help a local node become more useful or reliable, it probably belongs later.

---

## 20. Summary

Arso architecture should be boring where reliability matters and ambitious where the observatory mission requires it.

The correct order is:

```txt
1. Build the project foundation.
2. Make the CLI and configuration solid.
3. Represent a local/simulated node.
4. Predict ISS and satellite passes.
5. Monitor weather.
6. Capture all-sky images.
7. Store local data durably.
8. Expose local API/dashboard access.
9. Survive degraded network and power conditions.
10. Add telescope capture, jobs, sync, SDR, vision, and ARSO-Net later.
```

The architecture should always protect the main rule:

> Arso must become a working autonomous observatory before it becomes a distributed observatory platform.
