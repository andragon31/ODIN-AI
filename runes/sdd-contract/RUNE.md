# SDD-Contract Rune

## Purpose

Define the technical interface (Contract) between components. This rune ensures that the "how" of communication is established and validated before any implementation logic is written, preventing integration mismatches.

## When to Use

Use this rune when:
- The orchestrator launches you into the **Contract** phase.
- Methodology is **PENTAKILL**.
- Designing an API, a gRPC service, or a Go Interface for a module.

## Workflow

### Step 1: Interface Identification

Read the **Domain Model** (`openspec/domain/`) to identify which Aggregates or Services need an interface. 
Determine the best format:
- **gRPC/Protobuf**: For internal services or high-performance Go-to-Go communication.
- **OpenAPI/JSON**: For RESTful APIs or external integrations.
- **Go Interfaces**: For internal library boundaries.

### Step 2: Define the Schema

Escribir el esquema técnico riguroso. 
- Use naming from the Ubiquitous Language.
- Include validation rules (e.g., `min_length`, `regex`).
- Define error responses.

### Step 3: Write Contract Artifact

```protobuf
// Example for Protobuf
syntax = "proto3";
package odin.{domain};

service {ServiceName} {
  rpc {Action}({Request}) returns ({Response});
}

message {Request} {
  string id = 1;
}
```

## Rules

- NO implementation logic (no loops, no calculations).
- Contracts MUST be versioned (start with v1).
- Every RPC/Endpoint MUST have a Success and at least two Error (Validation/NotFound) responses documented.
- The contract is FIXED once the `contract` phase is completed. Changes require a new delta.

## Persistence

- **engram**: Save as `sdd/{change-name}/contract`
- **openspec**: Write to `openspec/contracts/{domain-name}.proto|yaml|go`
