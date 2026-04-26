---
title: Per-screen prototype sub-passes
status: superseded by 0070
date: 2026-04-10
---

## Context

Each screen is a learning opportunity about bubbletea
idioms. Lessons from building the sidebar inform the message list.
Incremental validation — each sub-pass produces a testable result.

## Decision

Pass 2.5b broken into 7 sub-passes, one screen at a
time: chrome shell, sidebar, message list, viewer, help popover,
status/toast, command mode.

## Consequences

The per-screen-prototype intent held; the specific seven-sub-pass
plan did not. Command mode dropped (ADR-0024); threading and
sidebar search were added as their own sub-passes when the screens
revealed them as distinct concerns. Superseded by ADR-0070, which
documents the realized sub-pass set.
