# SDD-Deploy Rune (Dvergar Pillar)

## Purpose

The **Dvergar** pillar represents the "Forge" where software becomes tangible. This rune guides the transition from verified code to deployed infrastructure. 

## Workflow: Consultative Inquiry Mode

Because deployment is highly dependent on user intent and environment, this rune does NOT execute automatically. It enters an **Inquiry Phase** to align with the user.

### Step 1: Detect Stack

Analyze the project structure to identify the technology stack (e.g., Go, Node.js, Python) and any existing infrastructure files (Dockerfiles, Terraform).

### Step 2: Proactive Inquiry (Dynamic Conversation)

**ODIN MUST NOT GUESS.** If the deployment target is not obvious, start a conversation.

**Standard Inquiry Questions:**
1.  **Destino (Target)**: ¿Dónde quieres desplegar? (Cloud Run, AWS, Heroku, Docker Local, etc.)
2.  **Infraestructura**: ¿Deseas que genere archivos de IAcC (Terraform, CloudFormation, K8s)?
3.  **CI/CD**: ¿Necesitas automatización con GitHub Actions o GitLab CI?
4.  **Estrategia**: ¿Prefieres un despliegue directo o algo más avanzado como Blue-Green?

### Step 3: Forge Infrastructure

Based on the conversation, generate the necessary artifacts:
- `infra_design.md`: A summary of the chosen infrastructure.
- Deployment files (e.g., `Dockerfile`, `docker-compose.yaml`, `deploy.sh`).
- CI/CD workflows.

## Rules

- **ALWAYS** ask before assuming a cloud provider.
- **NEVER** store credentials in the generated files. Use environment variables.
- Match the naming conventions defined in `openspec/domain/`.
- If the user choice is complex, create an `infra_design.md` for approval before generating code.

## Return Envelope

Return to the orchestrator:
- Summary of the chosen deployment path.
- List of infrastructure files created.
- Instructions for the user to execute the deployment (if manual).
