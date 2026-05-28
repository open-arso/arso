# Arso Vision

**Arso** stands for **Autonomous Remote Space Observatory**.

Arso is an open-source autonomous remote space and weather observatory platform. Its goal is to help makers, students, software engineers, researchers, amateur astronomers, and curious communities build connected observatory nodes capable of sky imaging, weather monitoring, satellite tracking, AI-assisted object recognition, and distributed data sharing.

The long-term ambition is to make remote sky observation more accessible, reproducible, programmable, and collaborative.

---

## 1. Mission

Arso aims to provide a complete open-source software and hardware foundation for building autonomous observatory nodes.

An Arso node should eventually be able to:

- observe the sky automatically;
- capture images of celestial objects and satellites;
- monitor local weather and environmental conditions;
- schedule observation jobs;
- expose its status and data through APIs and a CLI;
- support AI-assisted analysis of captured images;
- participate in a distributed network of observatories;
- be reproducible by individuals, schools, clubs, and research-oriented communities.

The project should be useful both as a real observatory system and as a learning platform for astronomy, embedded systems, distributed systems, computer vision, remote operations, and space-adjacent engineering.

---

## 2. North Star

The North Star of Arso is:

> Make it possible for anyone to build, operate, and connect an autonomous observatory node using open software, affordable hardware, and documented engineering practices.

The project should not merely be a toy telescope controller. It should become a serious learning and experimentation platform for people who want to understand the technical foundations behind remote observation, automation, satellite tracking, sensor networks, and distributed infrastructure.

---

## 3. Why Arso Exists

Modern space and astronomy systems are multidisciplinary. They require knowledge of:

- software engineering;
- hardware integration;
- networking;
- distributed systems;
- image processing;
- observability;
- automation;
- scheduling;
- geospatial calculations;
- weather constraints;
- data storage;
- operational reliability.

Most learning projects cover only one small part of this stack. Arso exists to connect these domains into one coherent project.

A useful Arso system should let someone learn by building:

- a command-line interface;
- a backend API;
- an edge agent;
- a hardware control loop;
- a camera capture pipeline;
- a weather monitoring system;
- a satellite tracking module;
- an image recognition service;
- a local dashboard;
- a distributed network of observatory nodes.

This makes Arso a strong long-term engineering project rather than a collection of unrelated experiments.

---

## 4. Target Users

Arso is designed for several types of users.

### 4.1 Makers and Hobbyists

People who want to build a Raspberry Pi-based observatory, connect a camera, monitor the sky, and automate simple observations.

### 4.2 Amateur Astronomers

People who want programmable tools for sky imaging, satellite pass prediction, object tracking, and local observatory automation.

### 4.3 Students and Learners

People who want a concrete project to learn about astronomy, embedded systems, distributed systems, computer vision, cloud infrastructure, and remote operations.

### 4.4 Clubs, Schools, and Communities

Groups that may want to deploy multiple observatory nodes, compare weather and sky conditions, and share captured data.

### 4.5 Researchers and Advanced Experimenters

People who may later want to use Arso nodes for distributed sky monitoring, environmental sensing, satellite observation, or citizen-science datasets.

---

## 5. Product Vision

In its mature form, Arso should provide:

- a documented open-source monorepo;
- a CLI named `arso`;
- an edge agent running on a Raspberry Pi or similar device;
- an API server for remote management and data access;
- a web dashboard;
- weather sensor integrations;
- camera and telescope mount integrations;
- satellite and celestial object tracking;
- image capture and storage;
- AI-assisted image analysis;
- observation scheduling;
- node health monitoring;
- secure remote operations;
- mesh-style data sharing between nodes;
- reproducible hardware profiles;
- clear documentation for contributors and builders.

The system should be modular enough to run locally at first, then evolve into a distributed observatory network.

---

## 6. Example Future Experience

A user should eventually be able to run:

```bash
arso node status
```

and see the state of their observatory:

```txt
Node: backyard-pi-001
Status: online
Camera: connected
Weather: safe
Cloud cover: 12%
Mount: tracking-capable
Storage: 68% free
Last capture: 2026-05-28 01:42:12
```

They should be able to search for a satellite pass:

```bash
arso find ISS
```

and get:

```txt
Target: ISS
Next visible pass: 2026-05-28 22:14:32 local time
Max elevation: 64°
Direction: WSW -> ENE
Duration: 5m 42s
Visibility: good
```

They should eventually be able to launch an observation:

```bash
arso observe moon --duration 10m
```

or follow a satellite:

```bash
arso follow ISS --camera live
```

The dashboard should show live node status, scheduled jobs, captured images, weather conditions, and image analysis results.

---

## 7. Guiding Principles

### 7.1 Working Hardware First

Arso should prioritize real-world operation over abstract architecture. A simple observatory that works is more valuable than a complex platform that only exists on diagrams.

### 7.2 Local First, Distributed Later

The first goal is a reliable local node. Distributed networking, mesh synchronization, and community registries should come only after a single node works well.

### 7.3 Simple Deployment First

Docker Compose should come before Kubernetes. A clear local setup should come before cloud-native complexity.

### 7.4 Clear Protocols Before Many Services

Arso should define stable data models and contracts before splitting logic into too many services.

### 7.5 Simulated Mode Is Required

Not everyone will have hardware. The project should support simulated camera, weather, and tracking modes so contributors can develop without a full observatory setup.

### 7.6 Reproducibility Matters

Arso should document hardware profiles, wiring, software setup, calibration steps, and operating procedures.

### 7.7 Useful Data Over Pretty UI

A beautiful dashboard is not enough. Arso should produce useful, structured, timestamped, and queryable data.

### 7.8 Community-Friendly by Default

The project should be welcoming to contributors, but technically serious. Documentation, issues, and contribution paths should make it easy to participate.

---

## 8. Non-Goals

Arso should avoid the following, especially in early versions:

- becoming a Kubernetes demo before a single node works;
- introducing Kafka or heavy event infrastructure too early;
- trying to support every telescope, camera, and sensor from day one;
- storing large AI model weights directly in the Git repository;
- promising scientific-grade measurements before calibration exists;
- building a social network before the observatory system is reliable;
- making the CLI, API, web app, and agent all duplicate business logic;
- becoming dependent on expensive proprietary hardware.

---

## 9. Technical Vision

Arso should use a layered architecture.

```txt
CLI / Web Dashboard
        |
        v
API Server
        |
        v
Scheduler / Job System
        |
        v
Edge Agent
        |
        v
Camera / Mount / Weather Sensors
```

Additional services can be added as the project matures:

```txt
Captured Images
      |
      v
Vision Service
      |
      v
Detections / Metadata / Reports
```

And later:

```txt
Arso Node A <----> Arso Node B <----> Arso Node C
        \              |              /
         \             |             /
          v            v            v
             Community Registry
```

The first implementation should keep this architecture simple. The boundaries should exist conceptually before they become operationally complex.

---

## 10. Open-Source Vision

Arso should be developed as an open-source project with:

- clear documentation;
- reproducible examples;
- beginner-friendly issues;
- technical design documents;
- transparent roadmap;
- permissive licensing;
- hardware build guides;
- documented APIs;
- automated tests;
- realistic contribution standards.

The project should not depend on one person’s private setup. A motivated contributor should eventually be able to build their own node by following the documentation.

---

## 11. Long-Term Ambition

The long-term ambition is to create a distributed network of Arso observatories.

Each node could contribute:

- local weather measurements;
- sky images;
- satellite observations;
- light pollution indicators;
- cloud coverage estimates;
- object detection metadata;
- observation availability windows.

Over time, this could become a community-driven observatory network useful for education, experimentation, and citizen science.

This ambition should not distract from the first milestone: one working autonomous node.

---

## 12. Success Definition

Arso is successful when:

- a new user can build or simulate a node;
- the node can capture images and collect weather data;
- the CLI and API can control the node;
- observation jobs can be scheduled and executed;
- captured data is stored and browsable;
- AI analysis can enrich captured images;
- documentation is good enough for others to reproduce the setup;
- multiple nodes can eventually share metadata safely and usefully.

The project should grow from a working local observatory into a distributed open observatory ecosystem.

---

## 13. Project Motto

> Build the node. Automate the sky. Share the data.
