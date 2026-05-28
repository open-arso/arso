# Arso Roadmap

**Arso** is an open-source autonomous remote space and weather observatory platform.

This roadmap focuses on building a real local observatory first, then making it reliable, then adding synchronization, radio reception, computer vision, and finally network-scale collaboration through **ARSO-Net**.

The guiding order is:

```txt
predict → measure → capture → store → record → survive → sync → receive → analyze → share
```

---

## Roadmap Philosophy

Arso should grow as an operational observatory, not as a premature distributed platform.

The project should prioritize:

1. **Satellite/ISS prediction as the first real feature**
2. **Reliable local sensing before advanced automation**
3. **All-sky capture before telescope capture**
4. **Local storage before event recording, sync, AI, or sharing**
5. **Event recording before AI interpretation**
6. **Degraded/offline operation before network synchronization**
7. **Useful datasets before computer vision**
8. **ARSO-Net only after a node can operate independently**

The first serious milestone is not mesh networking, AI, or a public registry.

The first serious milestone is:

```txt
A local Arso node can predict satellite passes, monitor weather, capture sky images,
store them locally with metadata, and recover safely from ordinary failures.
```

---

## Priority Model

The functional priority of the project is:

| Priority | Capability | Rationale |
|---:|---|---|
| 1 | Predict satellite and ISS passes | First useful space feature; requires no camera, no weather station, no telescope, no SDR, and no motorized mount. |
| 2 | Monitor local weather | Weather determines observation quality and hardware safety once capture begins. |
| 3 | Capture all-sky images | Produces the first real observatory data with simple hardware. |
| 4 | Store data locally | Required before event recording, sync, AI, or sharing. |
| 5 | Capture telescope images | Adds target-specific observation after the all-sky path works. |
| 6 | Record sky events | Builds structured observation history from images and sensor data. |
| 7 | Operate in degraded network and power conditions | Makes the node trustworthy when unattended. |
| 8 | Synchronize data later | Sync is useful only after local storage and recovery are reliable. |
| 9 | Receive selected radio signals using SDR | Extends Arso beyond optical observation. |
| 10 | Detect objects and anomalies using computer vision | Requires accumulated images/events to be genuinely useful. |
| 11 | Share metadata and observations with ARSO-Net | Comes last because public/distributed sharing requires stable local semantics. |

The key correction is that **satellite/ISS prediction should be the first real feature** because it is hardware-light and immediately gives Arso a space-oriented identity.

However, prediction should not come before the minimum foundations:

```txt
repo → CLI → config → simulated node → satellite prediction
```

Without configuration, observer location, and time handling, the prediction feature becomes a throwaway script instead of part of the platform.

Local storage still needs to come before serious event recording, synchronization, AI, or ARSO-Net. Degraded operation still needs to come before synchronization.

---

## Version Overview

| Version | Theme | Main Goal |
|---|---|---|
| v0.1 | Project foundation | Repository, documentation, local dev conventions, simulated mode design |
| v0.2 | CLI and local configuration | First `arso` commands and configuration profiles |
| v0.3 | Edge agent and simulated node | Local node process, health, capabilities, simulated camera/weather |
| v0.4 | Satellite and ISS pass prediction MVP | Predict visible passes from configured observer location |
| v0.5 | Weather monitoring MVP | Current weather, weather history, safe-to-observe baseline |
| v0.6 | All-sky image capture MVP | Capture all-sky images and attach metadata |
| v0.7 | Local storage and metadata model | Store observations, weather, predictions, images, and events locally |
| v0.8 | Local API and dashboard MVP | Browse node status, captures, weather, passes, and events |
| v0.9 | Degraded operation MVP | Offline-first behavior, recovery after restart, basic power/network resilience |
| v1.0 | Stable all-sky node release | Reproducible single-node all-sky observatory |
| v1.1 | Telescope capture MVP | Targeted telescope/camera capture without advanced mount automation |
| v1.2 | Observation jobs and scheduler | Plan captures using weather and pass windows |
| v1.3 | Remote operation hardening | Secure remote access, logs, node heartbeat, safe remote actions |
| v1.4 | Data synchronization | Sync local data to another machine/server when network is available |
| v1.5 | SDR reception MVP | Receive selected radio signals and store signal metadata |
| v1.6 | Formal hardware profiles | Reproducible hardware profiles, BOMs, wiring, validation checklists |
| v2.0 | Computer vision and anomaly detection | Analyze accumulated images/events with practical CV models |
| v2.1 | ARSO-Net metadata sharing MVP | Share selected metadata between nodes with privacy controls |
| v2.2 | Community node registry | Optional public/private registry of nodes and capabilities |
| v2.3 | Distributed observation campaigns | Coordinate observations across multiple ARSO-Net nodes |

---

# v0.x — Local Prototype Phase

The v0 phase proves that Arso can work as a local observatory.

The goal is to produce real data early, while still supporting simulated mode for contributors without hardware.

---

## v0.1 — Project Foundation

### Goal

Create the open-source foundation for `openarso/arso`.

### Deliverables

- Monorepo initialized
- `README.md`
- `LICENSE`
- `docs/vision.md`
- `docs/roadmap.md`
- `docs/architecture.md`
- `docs/cahier-des-charges.md`
- Basic repository structure
- Initial issue templates
- Initial pull request template
- Basic `Makefile` or task runner
- Local development documentation
- Simulated mode requirements documented

### Suggested Structure

```txt
arso/
├── apps/
│   ├── cli/
│   ├── api/
│   ├── web/
│   └── edge-agent/
├── packages/
│   ├── protocol/
│   ├── sdk/
│   └── common/
├── services/
│   ├── tracker/
│   ├── weather/
│   ├── vision/
│   └── sdr/
├── infra/
│   ├── compose/
│   └── docker/
├── docs/
└── README.md
```

### Success Criteria

- A contributor can understand the project structure quickly.
- The intended local-first architecture is clear.
- The first development commands are documented.

### Avoid For Now

- Multiple repositories
- Kubernetes
- Kafka
- Public registry
- AI model work
- Mesh networking

---

## v0.2 — CLI and Local Configuration

### Goal

Create the first usable `arso` CLI and configuration model.

The CLI should become the primary operator interface.

### Deliverables

- CLI application
- `arso version`
- `arso config init`
- `arso config show`
- Local configuration file
- Observer location configuration
- Local/simulated profile support
- JSON output option for automation

### Initial Commands

```bash
arso version
arso config init
arso config show
arso node status
arso weather current
arso find ISS
arso capture allsky
```

### Success Criteria

- CLI builds locally.
- CLI can load configuration.
- CLI can call mocked or simulated services.
- Observer location can be configured for pass prediction.

### Avoid For Now

- Direct hardware control in the CLI
- Complex authentication
- Too many flags before real behavior exists

---

## v0.3 — Edge Agent and Simulated Node

### Goal

Create the local agent that represents one observatory node.

### Deliverables

- `arso-agent`
- Health endpoint
- Node status endpoint
- Capability model
- Simulated camera
- Simulated weather provider
- Local logs
- Agent configuration

### Example Capability Model

```json
{
  "nodeId": "local-node-001",
  "capabilities": {
    "weather": true,
    "satellitePrediction": true,
    "allSkyCamera": true,
    "telescopeCamera": false,
    "eventRecording": false,
    "sdr": false,
    "computerVision": false,
    "sync": false,
    "arsoNet": false
  }
}
```

### Success Criteria

- Agent runs on a local Linux machine.
- CLI can query agent status.
- Simulated node works without hardware.
- Capabilities are explicit.

### Avoid For Now

- Complex hardware abstraction
- Remote multi-user control
- Mesh networking

---

## v0.4 — Satellite and ISS Pass Prediction MVP

### Goal

Predict visible satellite passes, starting with the ISS.

This gives Arso a strong early space-oriented feature before complex camera/telescope control.

### Deliverables

- Observer location model
- TLE ingestion or provider integration
- Satellite target model
- ISS pass prediction
- Pass list command
- Next pass command
- Azimuth/elevation time-series output
- Dry-run follow plan

### CLI Examples

```bash
arso find ISS
arso pass next --target ISS
arso pass list --target ISS --next 24h
arso follow ISS --dry-run
```

### Success Criteria

- User can get next ISS pass for configured location.
- Output includes start, peak, end, direction, duration, and max elevation.
- Dry-run follow output can generate azimuth/elevation points.
- Feature works without mount hardware.
- Feature works on a Raspberry Pi.
- Prediction logic is separated from hardware movement.

### Avoid For Now

- Precision motor tracking
- Guaranteed visual ISS tracking
- Radio communication claims
- Multi-node pass coordination

---


## v0.5 — Weather Monitoring MVP

### Goal

Monitor local weather and environmental conditions relevant to observation.

Weather should come early because it affects everything else.

### Deliverables

- Weather data model
- Simulated weather provider
- Public weather API provider
- Optional local sensor provider
- Current weather command
- Weather history persistence
- Basic safe-to-observe decision

### Weather Fields

- Temperature
- Humidity
- Pressure
- Wind speed
- Wind direction
- Cloud cover
- Rain probability or rain detection
- Dew point
- Visibility
- Provider/source
- Timestamp

### CLI Examples

```bash
arso weather current
arso weather history --last 24h
arso weather safe-to-observe
```

### Success Criteria

- Current weather can be displayed from CLI.
- Weather measurements can be stored locally.
- `safeToObserve` can be computed from simple rules.
- Weather metadata can later be attached to captures.

### Avoid For Now

- Complex weather ML
- Many providers
- Fully automated hardware protection without validation

---


## v0.6 — All-Sky Image Capture MVP

### Goal

Capture all-sky images and attach basic metadata.

All-sky capture should come before telescope capture because it is simpler, more robust, and useful for event recording.

### Deliverables

- All-sky camera interface
- Simulated all-sky capture
- Real camera adapter for one common camera
- Capture command
- Capture metadata
- Local image file output
- Basic image preview path
- Weather metadata attached to capture when available
- Pass prediction metadata attached to capture when relevant

### CLI Examples

```bash
arso capture allsky
arso capture allsky --exposure 100ms
arso captures list
arso captures show <capture-id>
```

### Capture Metadata Example

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

### Success Criteria

- A user can trigger an all-sky capture.
- Image is written locally.
- Metadata is produced.
- Capture can include weather metadata.
- Capture can be linked to a satellite pass when relevant.
- Simulated capture exists.

### Avoid For Now

- Telescope mount automation
- Astrophotography stacking
- AI detection
- Scientific calibration claims

---


## v0.7 — Local Storage and Metadata Model

### Goal

Define and implement durable local storage before data volume grows.

This must come before serious event recording, synchronization, AI, or ARSO-Net.

### Deliverables

- Local database or structured file store
- Observation metadata model
- Weather measurement persistence
- Satellite pass prediction persistence
- Capture metadata model
- Event metadata model
- Local object/file storage layout
- Backup/export command skeleton
- Basic schema/versioning approach

### Data Categories

- Node metadata
- Observer location
- Weather measurements
- Satellite pass predictions
- All-sky captures
- Telescope captures
- Sky events
- System events
- Power/network events
- SDR captures
- Analysis results

### CLI Examples

```bash
arso data stats
arso data export --format json
arso data list captures
arso data list events
```

### Success Criteria

- Data survives process restart.
- Metadata and files are linked consistently.
- Storage format is documented.
- Local data can be inspected without a cloud service.
- The system has a clear migration/versioning strategy.

### Avoid For Now

- Distributed sync
- Public sharing
- Complex database migrations before schema stabilizes

---


## v0.8 — Local API and Dashboard MVP

### Goal

Expose local observatory data through an API and simple dashboard.

The API/dashboard should support local operation, not become a SaaS platform.

### Deliverables

- API application
- Health endpoint
- Node status endpoints
- Weather endpoints
- Satellite pass endpoints
- Capture endpoints
- Event endpoints
- Basic dashboard
- Local Docker Compose setup

### Initial Endpoints

```http
GET    /health
GET    /api/v1/node/status
GET    /api/v1/node/capabilities
GET    /api/v1/weather/current
GET    /api/v1/weather/history
GET    /api/v1/satellites/{target}/passes
POST   /api/v1/captures/allsky
GET    /api/v1/captures
GET    /api/v1/captures/{captureId}
GET    /api/v1/events
GET    /api/v1/events/{eventId}
```

### Dashboard Pages

- Node status
- Weather
- Satellite passes
- Captures
- Events
- Local storage status

### Success Criteria

- Browser can show node status.
- Browser can list captures and weather.
- API is documented.
- CLI can target API instead of only agent.

### Avoid For Now

- Public accounts
- Multi-tenant assumptions
- Complex user roles
- Social features

---

## v0.9 — Degraded Operation MVP

### Goal

Make the node robust when network or power conditions are imperfect.

This comes before synchronization because sync is useless if the local node loses or corrupts data during failures.

### Deliverables

- Offline-first local writes
- Restart recovery
- Agent watchdog guidance
- Job/capture resume rules
- Storage integrity checks
- Network unavailable state
- Power event logging, where supported
- Safe shutdown documentation
- Basic diagnostics command

### CLI Examples

```bash
arso diagnostics
arso data verify
arso node recovery-status
```

### Success Criteria

- Node can keep storing local data without network.
- Node can recover cleanly after restart.
- Interrupted captures/jobs are marked clearly.
- Local data integrity can be checked.
- Network loss does not break local observation.

### Avoid For Now

- Complex distributed consensus
- Multi-node failover
- Solar/battery optimization
- Hardware-specific UPS assumptions

---

# v1.x — Reliable Single-Node Observatory Phase

The v1 phase turns Arso into a reproducible observatory node that other people can install and operate.

---

## v1.0 — Stable All-Sky Node Release

### Goal

Release the first stable local Arso node.

### Deliverables

- Stable CLI subset
- Stable agent subset
- Stable local API subset
- Basic dashboard
- Weather monitoring
- ISS/satellite pass prediction
- All-sky capture
- Local storage
- Degraded operation behavior
- Installation guide
- Troubleshooting guide
- Minimal hardware guide

### Definition of Done

- Fresh install has been tested.
- Simulated mode works.
- At least one real all-sky camera setup is documented.
- Local data survives restart.
- CLI/API/dashboard can inspect captures, weather, passes, and events.
- The system works without ARSO-Net or cloud services.

---

## v1.1 — Telescope Capture MVP

### Goal

Add targeted telescope image capture.

This milestone is about capture first, not full autonomous pointing.

### Deliverables

- Telescope camera capture interface
- Telescope capture metadata
- Target metadata
- Manual target selection
- Optional mount adapter skeleton
- Telescope capture command
- Storage integration
- Dashboard support

### CLI Examples

```bash
arso capture telescope --target moon
arso capture telescope --target jupiter
arso captures list --type telescope
```

### Success Criteria

- Telescope image can be captured and stored.
- Telescope captures use the same local metadata/storage model.
- Target information can be associated with the capture.
- System still works without telescope hardware.

### Avoid For Now

- Full autonomous mount control
- Precision guiding
- Advanced stacking
- Universal telescope support

---

## v1.2 — Observation Jobs and Scheduler

### Goal

Plan and execute observation jobs using weather and pass windows.

### Deliverables

- Job model
- Job state machine
- Scheduler
- One-shot capture jobs
- Recurring weather sampling jobs
- Satellite pass capture jobs
- Job logs
- Job cancellation

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

### CLI Examples

```bash
arso job create allsky --every 5m
arso job create satellite-pass --target ISS
arso job list
arso job logs <job-id>
arso job cancel <job-id>
```

### Success Criteria

- Jobs are stored locally.
- Jobs can execute through the agent.
- Weather rules can prevent jobs from running.
- Interrupted jobs are recoverable or clearly marked.

### Avoid For Now

- Complex optimization
- Multi-node scheduling
- Public campaign features

---

## v1.3 — Remote Operation Hardening

### Goal

Operate a private Arso node remotely and safely.

### Deliverables

- Authentication for remote API access
- Node heartbeat
- Remote logs
- Remote status
- Remote capture trigger
- Safe remote command model
- Basic authorization
- Secure configuration guidelines

### CLI Examples

```bash
arso node register
arso node status <node-id>
arso node logs <node-id>
arso capture allsky --node <node-id>
```

### Success Criteria

- User can inspect a node remotely.
- User can trigger safe remote actions.
- Unsafe hardware actions require explicit safeguards.
- Credentials are not logged.
- Precise location is not exposed publicly by default.

### Avoid For Now

- Public SaaS
- Public remote control
- Fine-grained enterprise permission model

---

## v1.4 — Data Synchronization

### Goal

Synchronize local observatory data to another machine or server when network is available.

This is not yet ARSO-Net. It is private/local-owner sync.

### Deliverables

- Sync status model
- Outbox/inbox model
- File sync strategy
- Metadata sync strategy
- Conflict rules
- Retry/backoff
- Resume after network loss
- Sync diagnostics

### Sync Targets

- Local server
- NAS
- Remote VPS
- S3-compatible storage
- Another private Arso instance

### CLI Examples

```bash
arso sync status
arso sync run
arso sync verify
```

### Success Criteria

- Local data remains source of truth.
- Sync can resume after interruption.
- Failed sync does not corrupt local data.
- User can choose what is synchronized.

### Avoid For Now

- Public ARSO-Net sharing
- Community registry
- Raw image sharing by default
- Complex peer-to-peer mesh

---

## v1.5 — SDR Reception MVP

### Goal

Receive selected radio signals using SDR hardware and store signal metadata.

This extends Arso beyond optical observation while still remaining local-first.

### Deliverables

- SDR capability model
- Supported SDR hardware profile for one device family
- Frequency configuration
- Recording command
- Signal metadata storage
- Basic waterfall/spectrum artifact storage, if feasible
- Safety/legal documentation note
- Dashboard/API listing for SDR captures

### CLI Examples

```bash
arso sdr devices
arso sdr listen --frequency 137.1MHz --duration 2m
arso sdr captures list
```

### Success Criteria

- Arso can detect or configure an SDR device.
- A selected frequency can be recorded locally.
- SDR metadata is stored with timestamp and configuration.
- SDR feature is optional and disabled by default.

### Avoid For Now

- Transmitting radio signals
- Decoding every protocol
- Legal claims about frequency use
- Complex DSP pipeline

---

## v1.6 — Formal Hardware Profiles

### Goal

Make common Arso builds reproducible.

Basic hardware notes should exist earlier, but this milestone formalizes supported profiles.

### Deliverables

- Hardware profile schema
- Simulated node profile
- Minimal all-sky node profile
- Weather node profile
- Telescope capture profile
- SDR receiver profile
- Wiring diagrams
- Bill of materials examples
- Validation checklist
- Known limitations per profile

### Example Profiles

```txt
simulated-node
minimal-allsky-pi
weather-allsky-node
telescope-capture-node
sdr-listener-node
```

### Success Criteria

- Users can select a profile and follow setup steps.
- Software capabilities map clearly to hardware capabilities.
- Unsupported hardware can still run in simulated mode.
- Hardware docs are honest about limitations.

### Avoid For Now

- Claiming universal support
- Expensive hardware as the default path
- Unsafe motor defaults

---

# v2.x — Analysis and Network Phase

The v2 phase starts only after a useful local node exists and enough data has been collected to make analysis and sharing meaningful.

---

## v2.0 — Computer Vision and Anomaly Detection

### Goal

Use accumulated images and events to detect useful objects, conditions, and anomalies.

AI comes here because it needs data. Earlier AI would mostly be demo logic.

### Deliverables

- Vision service
- Image quality analysis
- Cloud/blur/overexposure detection
- Moon/Sun/star-field classification
- Satellite/aircraft trail candidate detection
- Event anomaly scoring
- Manual review workflow
- Analysis metadata storage
- Re-analysis command

### CLI Examples

```bash
arso image analyze <capture-id>
arso image analyze --last 24h
arso events review
```

### Success Criteria

- Analysis enriches existing captures/events.
- Results are stored as metadata.
- System does not require GPU by default.
- Results are presented as probabilities/candidates, not scientific certainty.

### Avoid For Now

- Scientific-grade claims
- Training custom models before dataset quality is known
- Mandatory cloud AI providers
- Automatic public anomaly claims

---

## v2.1 — ARSO-Net Metadata Sharing MVP

### Goal

Share selected metadata and observations with a wider network called **ARSO-Net**.

This is the first public/distributed sharing milestone.

### Deliverables

- ARSO-Net identity model
- Public/private sharing settings
- Metadata export format
- Metadata import format
- Privacy boundary model
- Node sharing policy
- Remote data labeling
- Opt-in synchronization

### Shareable Data Types

- Approximate node region
- Node capabilities
- Weather summaries
- Observation summaries
- Capture metadata
- Event metadata
- Analysis metadata
- SDR metadata summaries

### Not Shared By Default

- Exact private location
- Raw images
- Credentials
- Personal information
- Hardware control access
- Private logs

### Success Criteria

- A node can export selected public metadata.
- A node can import ARSO-Net metadata.
- Remote data is clearly marked as remote.
- User can disable sharing completely.

### Avoid For Now

- Public remote control
- Exact home location exposure
- Blockchain/token systems
- Mandatory central service

---

## v2.2 — Community Node Registry

### Goal

Allow users to optionally publish discoverable ARSO-Net node profiles.

The registry comes after metadata sharing because the project should know what it is registering before building discovery around it.

### Deliverables

- Public/private node profile
- Node capability registry
- Optional public node map
- Approximate location controls
- Last-seen status
- Moderation model
- Registry API
- Registry dashboard view

### Registry Fields

- Node name
- Visibility
- Approximate region
- Hardware profile
- Capabilities
- Shared data types
- Last seen timestamp
- Contact/project link, optional

### Success Criteria

- Users can register public or private nodes.
- Public nodes can be discovered by capability and region.
- Exact location is never required for public registration.
- Local operation does not depend on the registry.

### Avoid For Now

- Social network features
- Public command/control
- Reputation economy
- Mandatory registration

---

## v2.3 — Distributed Observation Campaigns

### Goal

Coordinate observations across multiple ARSO-Net nodes.

### Deliverables

- Campaign model
- Campaign metadata
- Multi-node observation planning
- Regional sky/weather comparison
- Aggregated event timelines
- Dataset export
- Education/research-friendly reports

### Example Use Cases

- Compare sky conditions across regions
- Observe the same satellite pass from different cities
- Build public meteor/event datasets
- Correlate weather and image quality
- Compare SDR observations from different locations

### Success Criteria

- Multiple nodes can contribute to a campaign.
- Data is comparable across nodes.
- Contributors control what they share.
- Campaign data can be exported.

### Avoid For Now

- Scientific claims without validation
- Centralized dependency for local operation
- Complex consensus protocols

---

# Cross-Cutting Tracks

## CLI Track

### v0

- Configuration
- Node status
- ISS/satellite pass commands
- Weather commands
- All-sky capture commands
- Local data inspection

### v1

- Telescope capture commands
- Job commands
- Remote node commands
- Sync commands
- SDR commands

### v2

- Image analysis commands
- ARSO-Net commands
- Registry commands
- Campaign commands

---

## Agent Track

### v0

- Simulated node
- Health and capabilities
- Satellite prediction integration
- Weather provider
- All-sky capture
- Local storage
- Degraded behavior

### v1

- Telescope capture
- Scheduler execution
- Remote operations
- Sync
- SDR integration
- Hardware profiles

### v2

- Vision analysis integration
- ARSO-Net sharing
- Campaign participation

---

## Data Track

### v0

- Observer location
- Pass predictions
- Weather measurements
- Captures
- Local metadata
- Events
- Integrity checks

### v1

- Telescope captures
- Jobs
- Sync states
- SDR captures
- Hardware profile metadata

### v2

- Analysis results
- ARSO-Net metadata
- Registry profiles
- Campaign datasets

---

## Hardware Track

### v0

- Raspberry Pi-friendly deployment target
- Simulated profile
- No external hardware required for satellite prediction
- Minimal all-sky camera path
- Basic weather source

### v1

- Telescope capture profile
- SDR receiver profile
- Formal hardware profiles
- Validation checklists

### v2

- Community profiles
- Advanced mounts
- Improved sensors
- Higher-quality observatory builds

---

## AI/Vision Track

### v0

- No AI as a core feature
- Prepare metadata and event models for later analysis

### v1

- Optional simple image quality heuristics may exist
- No mandatory AI service

### v2

- Computer vision service
- Object and anomaly candidates
- Dataset building
- Re-analysis workflows

---

## ARSO-Net Track

### v0

- No ARSO-Net
- Local-first data model only

### v1

- Private sync only
- No public registry

### v2

- Metadata sharing
- Community registry
- Distributed campaigns

---

# Non-Goals for Early Versions

Arso should avoid the following before v2:

- AI-first product direction
- Public community registry
- Mesh networking
- Distributed consensus
- Public raw image sharing
- Public remote node control
- Exact location sharing
- Kubernetes as a requirement
- Kafka as a requirement
- Custom model training
- Supporting every telescope/camera/SDR
- Scientific-grade claims without calibration

---

# Immediate Next Steps

1. Keep `openarso/arso` as a monorepo.
2. Add the documentation set:
   - `docs/vision.md`
   - `docs/roadmap.md`
   - `docs/architecture.md`
   - `docs/cahier-des-charges.md`
3. Implement the first CLI command:

```bash
arso version
```

4. Implement configuration:

```bash
arso config init
arso config show
```

5. Implement simulated node status:

```bash
arso node status
```

6. Then build the first real functional path:

```txt
find ISS → weather current → capture allsky → store metadata locally
```

That path should remain the early north star.
