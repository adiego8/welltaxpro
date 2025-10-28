# WellTaxPro - Multi-Tenant Tax Admin Platform

## Project Overview

**WellTaxPro** is a provider-agnostic, white-label admin system designed to scale for any accounting firm. MyWellTax serves as the pilot implementation, working with an accounting firm to validate the platform's capabilities.

## Working Philosophy

### Collaboration & Critical Thinking
This project is led by a **critical thinker** who values:
- **Constructive feedback** on all development decisions and architectural choices
- **Honest opinions** about ideas, even when they challenge existing approaches
- **Alternative perspectives** from team members, stakeholders, and technical advisors
- **Evidence-based decision making** over assumptions

When contributing to WellTaxPro:
- **Challenge assumptions**: If something doesn't make sense, speak up
- **Propose alternatives**: Better ideas can come from anywhere
- **Provide context**: Explain the "why" behind your feedback
- **Be direct**: Honest, respectful critique is more valuable than agreement for its own sake
- **Think long-term**: Consider scalability, maintainability, and future implications

The goal is to build the **best possible platform**, not to validate existing ideas. Your critical analysis and honest feedback are essential to achieving this.

---

## Vision

WellTaxPro is a provider-agnostic, white-label admin system for accounting firms that:
- Handles tax document processing, client management, and billing
- Integrates with any existing tax software platform through adapters
- Maintains IRS-compliant security and audit controls
- Scales from small practices to enterprise firms
- Offers flexible deployment models (SaaS, dedicated, self-hosted)

**Core Value Proposition**: Accounting firms get a modern admin layer without abandoning their existing tax software investments.

---

## Architecture Principles

1. **Tenant isolation first**: Every design decision must preserve per-tenant data boundaries (RLS + per-tenant KMS keys)
2. **Provider abstraction over direct coupling**: Never hardcode to a specific tax platform's schema
3. **IRS-aligned compliance by default**: Security, audit, and privacy controls are non-negotiable table stakes
4. **Capabilities over CRUD**: Expose business operations, not database tables
5. **Idempotency everywhere**: All state-changing operations must be safely retryable
6. **Observable by design**: Every adapter, sync job, and API call must emit structured telemetry

---
