# ADR-PC-021: CI/CD & image build/publish

- **Status:** Proposed — stub pending Architect ratification
- **Date:** 2026-07-10
- **Applies to:** Platform Core (`pc-*`) services and product (`{xx}-`) services
- **Source of truth:** [Service Registry](../prd/service-registry.md) · [Capability Document](../prd/Medhen-Platform-Capability-Document.md)

## Context

Deployments must be reproducible and provenance-tracked; moving tags cause drift.

## Decision

Images are built only from signed-off code (push to `main` after a layer sign-off, or a release tag). Tags are immutable (long git SHA, semver on tags); there is no `latest`. Images are pushed to Zot and deploys pin the digest. Per-PR CI runs quality gates but does not publish.

## Consequences

Enforced by the quality system (IS-AQG-001) and the shared test/CI scaffold; reviews and gates check conformance. This is a stub capturing the platform's stated convention — expand with alternatives considered and full rationale before ratifying.
