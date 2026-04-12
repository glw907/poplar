---
title: Mock backend for prototype
status: accepted
date: 2026-04-10
---

## Context

Enables the prototype (Pass 2.5b) without backend
dependencies. Useful long-term for development, testing, and demos.
Pass 3 swaps mock for real JMAP adapter — no throwaway code.

## Decision

`internal/mail/mock.go` implements `mail.Backend`
with hardcoded data. Stays in the codebase permanently.

## Consequences

No follow-on notes recorded.
