---
applyTo: "**/Dockerfile,**/Dockerfile.*,**/*.dockerfile,**/docker-compose*.yml,**/docker-compose*.yaml,**/compose*.yml,**/compose*.yaml"
description: "Best practices for creating optimized, secure, and efficient Docker images and managing containers. Covers multi-stage builds, image layer optimization, security scanning, and runtime best practices."
---

# Vite/SPA Build-Time Environment Variables (Critical)

**For Vite, React, and other SPA frameworks:**

- Environment variables (e.g., `VITE_API_URL`) are injected at build time, not at runtime.
- Always pass required variables as Docker build arguments (e.g., `--build-arg VITE_API_URL=...`).
- Do NOT expect runtime environment variables or ConfigMaps to affect the built bundle.
- Changing pod/container env vars after build will NOT update the frontend's API endpoints.
- Reference: [Vite Environment Variables and Modes](https://vitejs.dev/guide/env-and-mode.html)

**CI/CD and Kubernetes:**

- Ensure your build pipeline or deployment script sets the correct API URL at build time.
- To verify, inspect the built JS bundle in the running container for the correct value.

**Realtime/SignalR build-time rule:**

- Do **not** set `VITE_SIGNALR_URL` by default. Prefer using the same base as `VITE_API_URL` unless there is a strong architectural reason to split endpoints.
- If `VITE_SIGNALR_URL` is set to a dedicated endpoint, require pre-deploy validation from outside the cluster:
  - DNS resolution succeeds
  - TCP reachability succeeds on the public port
  - `POST <signalr-url>/api/hubs/<hub>/negotiate?negotiateVersion=1` returns `200`
  - Browser/client transport connection succeeds (WebSocket or negotiated fallback)
- If dedicated realtime endpoint validation fails, block deployment and fall back to API-hosted hub URL.

**Why:**
If the wrong value is set at build time, the deployed SPA will call the wrong backend, resulting in failed API calls and broken features. This is a common pitfall with Vite and Dockerized SPAs.

**Always treat Vite env vars as build-time only.**

# Docker and Container Instructions

Follow container security best practices and ISE Building Containers guidance.

**IMPORTANT**: Use the `iseplaybook` MCP server to get the latest container best practices. Use `context7` MCP server for Docker documentation. Do not assume—verify current guidance.

## When to Apply

This instruction applies when:

- Creating or modifying Dockerfiles
- Configuring docker-compose files
- Building container images
- Deploying containerized applications
- Reviewing container security and performance

## Core Principles of Containerization

### **1. Immutability**

- **Principle:** Once a container image is built, it should not change. Any changes should result in a new image.
- **Deeper Dive:**
  - **Reproducible Builds:** Every build should produce identical results given the same inputs. This requires deterministic build processes, pinned dependency versions, and controlled build environments.
  - **Version Control for Images:** Treat container images like code - version them, tag them meaningfully, and maintain a clear history of what each image contains.
  - **Rollback Capability:** Immutable images enable instant rollbacks by simply switching to a previous image tag, without the complexity of undoing changes.
  - **Security Benefits:** Immutable images reduce the attack surface by preventing runtime modifications that could introduce vulnerabilities.
- **Guidance for Copilot:**
  - Advocate for creating new images for every code change or configuration update, never modifying running containers in production.
  - Recommend using semantic versioning for image tags (e.g., `v1.2.3`, `latest` for development only).
  - Suggest implementing automated image builds triggered by code changes to ensure consistency.
  - Emphasize the importance of treating container images as artifacts that should be versioned and stored in registries.
- **Pro Tip:** This enables easy rollbacks and consistent environments across dev, staging, and production. Immutable images are the foundation of reliable deployments.

### **2. Portability**

- **Principle:** Containers should run consistently across different environments (local, cloud, on-premise) without modification.
- **Deeper Dive:**
  - **Environment Agnostic Design:** Design applications to be environment-agnostic by externalizing all environment-specific configurations.
  - **Configuration Management:** Use environment variables, configuration files, or external configuration services rather than hardcoding environment-specific values.
  - **Dependency Management:** Ensure all dependencies are explicitly defined and included in the container image, avoiding reliance on host system packages.
  - **Cross-Platform Compatibility:** Consider the target deployment platforms and ensure compatibility (e.g., ARM vs x86, different Linux distributions).
- **Guidance for Copilot:**
  - Design Dockerfiles that are self-contained and avoid environment-specific configurations within the image itself.
  - Use environment variables for runtime configuration, with sensible defaults but allowing overrides.
  - Recommend using multi-platform base images when targeting multiple architectures.
  - Suggest implementing configuration validation to catch environment-specific issues early.
- **Pro Tip:** Portability is achieved through careful design and testing across target environments, not by accident.

### **3. Isolation**

- **Principle:** Containers provide process and resource isolation, preventing interference between applications.
- **Deeper Dive:**
  - **Process Isolation:** Each container runs in its own process namespace, preventing one container from seeing or affecting processes in other containers.
  - **Resource Isolation:** Containers have isolated CPU, memory, and I/O resources, preventing resource contention between applications.
  - **Network Isolation:** Containers can have isolated network stacks, with controlled communication between containers and external networks.
  - **Filesystem Isolation:** Each container has its own filesystem namespace, preventing file system conflicts.
- **Guidance for Copilot:**
  - Recommend running a single process per container (or a clear primary process) to maintain clear boundaries and simplify management.
  - Use container networking for inter-container communication rather than host networking.
  - Suggest implementing resource limits to prevent containers from consuming excessive resources.
  - Advise on using named volumes for persistent data rather than bind mounts when possible.
- **Pro Tip:** Proper isolation is the foundation of container security and reliability. Don't break isolation for convenience.

### **4. Efficiency & Small Images**

- **Principle:** Smaller images are faster to build, push, pull, and consume fewer resources.
- **Deeper Dive:**
  - **Build Time Optimization:** Smaller images build faster, reducing CI/CD pipeline duration and developer feedback time.
  - **Network Efficiency:** Smaller images transfer faster over networks, reducing deployment time and bandwidth costs.
  - **Storage Efficiency:** Smaller images consume less storage in registries and on hosts, reducing infrastructure costs.
  - **Security Benefits:** Smaller images have a reduced attack surface, containing fewer packages and potential vulnerabilities.
- **Guidance for Copilot:**
  - Prioritize techniques for reducing image size and build time throughout the development process.
  - Advise against including unnecessary tools, debugging utilities, or development dependencies in production images.
  - Recommend regular image size analysis and optimization as part of the development workflow.
  - Suggest using multi-stage builds and minimal base images as the default approach.
- **Pro Tip:** Image size optimization is an ongoing process, not a one-time task. Regularly review and optimize your images.

## Dockerfile Best Practices

### **1. Multi-Stage Builds (The Golden Rule)**

- **Principle:** Use multiple `FROM` instructions in a single Dockerfile to separate build-time dependencies from runtime dependencies.
- **Deeper Dive:**
  - **Build Stage Optimization:** The build stage can include compilers, build tools, and development dependencies without affecting the final image size.
  - **Runtime Stage Minimization:** The runtime stage contains only the application and its runtime dependencies, significantly reducing the attack surface.
  - **Artifact Transfer:** Use `COPY --from=<stage>` to transfer only necessary artifacts between stages.
  - **Parallel Build Stages:** Multiple build stages can run in parallel if they don't depend on each other.
- **Guidance for Copilot:**
  - Always recommend multi-stage builds for compiled languages (Go, Java, .NET, C++) and even for Node.js/Python where build tools are heavy.
  - Suggest naming build stages descriptively (e.g., `AS build`, `AS test`, `AS production`) for clarity.
  - Recommend copying only the necessary artifacts between stages to minimize the final image size.
  - Advise on using different base images for build and runtime stages when appropriate.
- **Benefit:** Significantly reduces final image size and attack surface.

#### Advanced Multi-Stage with Testing

````dockerfile
#### Advanced Multi-Stage with Testing

```dockerfile
# Stage 1: Dependencies
FROM node:22-alpine AS deps
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production && npm cache clean --force

# Stage 2: Build
FROM node:22-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Stage 3: Test
FROM build AS test
RUN npm run test
RUN npm run lint

# Stage 4: Production
FROM node:22-alpine AS production
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY --from=build /app/dist ./dist
COPY --from=build /app/package*.json ./
USER node
EXPOSE 3000
CMD ["node", "dist/main.js"]
````

#### .NET Example

> **Breaking Change (.NET 10):** .NET 10 container images default to **Ubuntu** (Noble Numbat), not Debian. **Debian images are no longer shipped for .NET 10.** If your workflow depends on Debian-based images, you must build custom images or remain on .NET 9. See: [Default .NET images use Ubuntu](https://learn.microsoft.com/dotnet/core/compatibility/containers/10.0/default-images-use-ubuntu).

**Simple Multi-Stage (Recommended for Most Projects)**

```dockerfile
# Build stage
FROM mcr.microsoft.com/dotnet/sdk:10.0 AS build
WORKDIR /src
COPY *.csproj .
RUN dotnet restore
COPY . .
RUN dotnet publish -c Release -o /app

# Runtime stage
FROM mcr.microsoft.com/dotnet/aspnet:10.0 AS runtime
WORKDIR /app
COPY --from=build /app .
EXPOSE 8080
ENTRYPOINT ["dotnet", "MyApp.dll"]
```

**Advanced Multi-Stage with Optimized NuGet Caching (Large Projects with Many Dependencies)**

```dockerfile
# Stage 1: Restore dependencies only (cached if .csproj files don't change)
FROM mcr.microsoft.com/dotnet/sdk:10.0 AS deps
WORKDIR /src
COPY src/*/**.csproj ./
RUN dotnet restore

# Stage 2: Build (reuses NuGet cache from deps stage)
FROM mcr.microsoft.com/dotnet/sdk:10.0 AS build
WORKDIR /src
COPY --from=deps /root/.nuget /root/.nuget
COPY src/*/**.csproj ./
# ⚠️ CRITICAL: Must restore in this stage to generate project.assets.json
RUN dotnet restore
COPY src/ .
RUN dotnet build -c Release --no-restore

# Stage 3: Publish (uses pre-built artifacts)
FROM build AS publish
WORKDIR /src
RUN dotnet publish -c Release -o /app --no-build

# Stage 4: Runtime
FROM mcr.microsoft.com/dotnet/aspnet:10.0 AS runtime
WORKDIR /app
COPY --from=publish /app .
EXPOSE 8080
ENTRYPOINT ["dotnet", "MyApp.dll"]
```

**CRITICAL RULES for .NET Multi-Stage Builds:**

1. **Never use `--no-restore` without first running `dotnet restore` in that stage** (or copying pre-populated obj/ directories)
   - `--no-restore` assumes `project.assets.json` already exists
   - Missing `project.assets.json` causes: `error NETSDK1004: Assets file not found`

2. **If caching NuGet packages across stages:**
   - Stage 1 (deps): `COPY *.csproj` → `RUN dotnet restore` (creates obj/project.assets.json)
   - Stage 2 (build): `COPY --from=deps /root/.nuget /root/.nuget` (copy package cache)
   - Stage 2 (build): `COPY *.csproj` → `RUN dotnet restore` (regenerate obj/project.assets.json using cached packages)
   - Stage 2 (build): `RUN dotnet build ... --no-restore` (now safe; assets file exists)

3. **Use `--no-build` in publish stage** when publishing from a pre-built project
   - This skips rebuilding and uses cached build artifacts (faster)
   - Only valid after a successful build in a prior stage

4. **Copy .csproj files before source code** to maximize layer caching
   - If only source changes, restore layer is reused
   - If .csproj changes, only restore layer is invalidated

````

### **2. Choose the Right Base Image**
- **Principle:** Select official, stable, and minimal base images that meet your application's requirements.
- **Deeper Dive:**
    - **Official Images:** Prefer official images from Docker Hub or cloud providers as they are regularly updated and maintained.
    - **Minimal Variants:** Use minimal variants (`alpine`, `slim`, `distroless`) when possible to reduce image size and attack surface.
    - **Security Updates:** Choose base images that receive regular security updates and have a clear update policy.
    - **Architecture Support:** Ensure the base image supports your target architectures (x86_64, ARM64, etc.).
- **Guidance for Copilot:**
    - Prefer Alpine variants when compatible; otherwise choose slim or distroless for glibc-dependent apps.
    - Use official language-specific images (e.g., `python:3.12-slim`, `openjdk:17-jre-slim`).
- Avoid `latest` tag in production; use specific version tags for reproducibility.
- Pin base images explicitly and review upgrade notes before bumping.
    - Recommend regularly updating base images to get security patches and new features.
- **Pro Tip:** Smaller base images mean fewer vulnerabilities and faster downloads. Always start with the smallest image that meets your needs.

#### Base Image Examples

```dockerfile
# ✅ Good - specific version with minimal variant
FROM node:22-alpine

# ✅ Good - slim variant for Python
FROM python:3.12-slim

# ✅ Better - distroless for maximum security
FROM gcr.io/distroless/nodejs22-debian12

# ❌ Avoid - mutable tags
FROM node:latest
FROM node:22
````

### **3. Optimize Image Layers**

- **Principle:** Each instruction in a Dockerfile creates a new layer. Leverage caching effectively to optimize build times and image size.
- **Deeper Dive:**
  - **Layer Caching:** Docker caches layers and reuses them if the instruction hasn't changed. Order instructions from least to most frequently changing.
  - **Layer Size:** Each layer adds to the final image size. Combine related commands to reduce the number of layers.
  - **Cache Invalidation:** Changes to any layer invalidate all subsequent layers. Place frequently changing content (like source code) near the end.
  - **Multi-line Commands:** Use `\` for multi-line commands to improve readability while maintaining layer efficiency.
- **Guidance for Copilot:**
  - Place frequently changing instructions (e.g., `COPY . .`) _after_ less frequently changing ones (e.g., `RUN npm ci`).
  - Combine `RUN` commands where possible to minimize layers (e.g., `RUN apt-get update && apt-get install -y ...`).
  - Clean up temporary files in the same `RUN` command (`rm -rf /var/lib/apt/lists/*`).
  - Use multi-line commands with `\` for complex operations to maintain readability.

#### Layer Optimization Examples

```dockerfile
# ❌ BAD: Multiple layers, inefficient caching
FROM ubuntu:20.04
RUN apt-get update
RUN apt-get install -y python3 python3-pip
RUN pip3 install flask
RUN apt-get clean
RUN rm -rf /var/lib/apt/lists/*

# ✅ GOOD: Optimized layers with proper cleanup
FROM ubuntu:20.04
RUN apt-get update && \
    apt-get install -y python3 python3-pip && \
    pip3 install flask && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
```

### **4. Use `.dockerignore` Effectively**

- **Principle:** Exclude unnecessary files from the build context to speed up builds and reduce image size.
- **Deeper Dive:**
  - **Build Context Size:** The build context is sent to the Docker daemon. Large contexts slow down builds and consume resources.
  - **Security:** Exclude sensitive files (like `.env`, `.git`) to prevent accidental inclusion in images.
  - **Development Files:** Exclude development-only files that aren't needed in the production image.
  - **Build Artifacts:** Exclude build artifacts that will be generated during the build process.
- **Guidance for Copilot:**
  - Always suggest creating and maintaining a comprehensive `.dockerignore` file.
  - Common exclusions: `.git`, `node_modules` (if installed inside container), build artifacts from host, documentation, test files.
  - Recommend reviewing the `.dockerignore` file regularly as the project evolves.
  - Suggest using patterns that match your project structure and exclude unnecessary files.

#### Comprehensive .dockerignore Example

```dockerignore
# Version control
.git*

# Dependencies (if installed in container)
node_modules
vendor
__pycache__

# Build artifacts
dist
build
*.o
*.so

# Development files
.env.*
*.log
coverage
.nyc_output

# IDE files
.vscode
.idea
*.swp
*.swo

# OS files
.DS_Store
Thumbs.db

# Documentation
*.md
docs/

# Test files
test/
tests/
spec/
__tests__/
```

### **5. Minimize `COPY` Instructions**

- **Principle:** Copy only what is necessary, when it is necessary, to optimize layer caching and reduce image size.
- **Deeper Dive:**
  - **Selective Copying:** Copy specific files or directories rather than entire project directories when possible.
  - **Layer Caching:** Each `COPY` instruction creates a new layer. Copy files that change together in the same instruction.
  - **Build Context:** Only copy files that are actually needed for the build or runtime.
  - **Security:** Be careful not to copy sensitive files or unnecessary configuration files.
- **Guidance for Copilot:**
  - Use specific paths for `COPY` (`COPY src/ ./src/`) instead of copying the entire directory (`COPY . .`) if only a subset is needed.
  - Copy dependency files (like `package.json`, `requirements.txt`) before copying source code to leverage layer caching.
  - Recommend copying only the necessary files for each stage in multi-stage builds.
  - Suggest using `.dockerignore` to exclude files that shouldn't be copied.

#### Optimized COPY Strategy

```dockerfile
# Copy dependency files first (for better caching)
COPY package*.json ./
RUN npm ci

# Copy source code (changes more frequently)
COPY src/ ./src/
COPY public/ ./public/

# Copy configuration files
COPY config/ ./config/

# Don't copy everything with COPY . .
```

### **6. Define Default User and Port**

- **Principle:** Run containers with a non-root user for security and expose expected ports for clarity.
- **Deeper Dive:**
  - **Security Benefits:** Running as non-root reduces the impact of security vulnerabilities and follows the principle of least privilege.
  - **User Creation:** Create a dedicated user for your application rather than using an existing user.
  - **Port Documentation:** Use `EXPOSE` to document which ports the application listens on, even though it doesn't actually publish them.
  - **Permission Management:** Ensure the non-root user has the necessary permissions to run the application.
- **Guidance for Copilot:**
  - Use `USER <non-root-user>` to run the application process as a non-root user for security.
- Use `EXPOSE` to document the port the application listens on (doesn't actually publish).
- Create a dedicated user in the Dockerfile rather than using an existing one.
- Ensure proper file permissions for the non-root user.
- Avoid privileged containers and `docker.sock` mounts in production.

#### Secure User Setup

```dockerfile
# Create a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set proper permissions
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose the application port
EXPOSE 8080

# Start the application
CMD ["node", "dist/main.js"]
```

### **7. Use `CMD` and `ENTRYPOINT` Correctly**

- **Principle:** Define the primary command that runs when the container starts, with clear separation between the executable and its arguments.
- **Deeper Dive:**
  - **`ENTRYPOINT`:** Defines the executable that will always run. Makes the container behave like a specific application.
  - **`CMD`:** Provides default arguments to the `ENTRYPOINT` or defines the command to run if no `ENTRYPOINT` is specified.
  - **Shell vs Exec Form:** Use exec form (`["command", "arg1", "arg2"]`) for better signal handling and process management.
  - **Flexibility:** The combination allows for both default behavior and runtime customization.
- **Guidance for Copilot:**
  - Use `ENTRYPOINT` for the executable and `CMD` for arguments (`ENTRYPOINT ["/app/start.sh"]`, `CMD ["--config", "prod.conf"]`).
  - For simple execution, `CMD ["executable", "param1"]` is often sufficient.
  - Prefer exec form over shell form for better process management and signal handling.
  - Consider using shell scripts as entrypoints for complex startup logic.
- **Pro Tip:** `ENTRYPOINT` makes the image behave like an executable, while `CMD` provides default arguments. This combination provides flexibility and clarity.

### **8. Environment Variables for Configuration**

- **Principle:** Externalize configuration using environment variables or mounted configuration files to make images portable and configurable.
- **Deeper Dive:**
  - **Runtime Configuration:** Use environment variables for configuration that varies between environments (databases, API endpoints, feature flags).
  - **Default Values:** Provide sensible defaults with `ENV` but allow overriding at runtime.
  - **Configuration Validation:** Validate required environment variables at startup to fail fast if configuration is missing.
  - **Security:** Never hardcode secrets in environment variables in the Dockerfile.
- **Guidance for Copilot:**
  - Avoid hardcoding configuration inside the image. Use `ENV` for default values, but allow overriding at runtime.
  - Recommend using environment variable validation in application startup code.
- Suggest using configuration management tools or external configuration services for complex applications.
- Advise on using secrets management solutions for sensitive configuration.
- If running on AKS with Workload Identity, ensure apps support `AZURE_FEDERATED_TOKEN_FILE`.

#### Environment Variable Best Practices

```dockerfile
# Set default values
ENV NODE_ENV=production
ENV PORT=3000
ENV LOG_LEVEL=info

# Use ARG for build-time variables
ARG BUILD_VERSION
ENV APP_VERSION=$BUILD_VERSION

# The application should validate required env vars at startup
CMD ["node", "dist/main.js"]
```

### Language-Specific Examples

##### Node.js Example

```dockerfile
# Build stage
FROM node:22-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
RUN npm run build

# Runtime stage
FROM node:22-alpine AS runtime
WORKDIR /app
COPY --from=build /app/dist ./dist
COPY --from=build /app/node_modules ./node_modules
USER node
EXPOSE 3000
CMD ["node", "dist/index.js"]
```

#### Python Example

```dockerfile
FROM python:3.12-slim AS base

# Set environment variables
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application
COPY . .

# Create non-root user
RUN useradd -m appuser
USER appuser

EXPOSE 8000
CMD ["python", "-m", "uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
```

## Container Security Best Practices

### **1. Non-Root User**

- **Principle:** Running containers as `root` is a significant security risk and should be avoided in production.
- **Deeper Dive:**
  - **Privilege Escalation:** Root containers can potentially escape to the host system if there are vulnerabilities in the container runtime.
  - **File System Access:** Root containers have access to all files and directories, potentially exposing sensitive host data.
  - **Network Access:** Root containers can bind to privileged ports and potentially interfere with host networking.
  - **Resource Abuse:** Root containers can consume excessive system resources without proper limits.
- **Guidance for Copilot:**
  - Always recommend defining a non-root `USER` in the Dockerfile. Create a dedicated user for your application.
  - Ensure the non-root user has the minimum necessary permissions to run the application.
  - Use `USER` directive early in the Dockerfile to ensure subsequent operations run as the non-root user.
  - Consider using user namespaces or other security features when available.

#### Secure User Creation

```dockerfile
# Create a dedicated user and group
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set proper ownership of application files
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Ensure the user can write to necessary directories
VOLUME ["/app/data"]
```

### **2. Minimal Base Images**

- **Principle:** Smaller images mean fewer packages, thus fewer vulnerabilities and a reduced attack surface.
- **Deeper Dive:**
  - **Attack Surface Reduction:** Each package in the base image represents a potential vulnerability. Fewer packages mean fewer potential attack vectors.
  - **Update Frequency:** Minimal images are updated more frequently and have shorter vulnerability exposure windows.
  - **Resource Efficiency:** Smaller images consume less storage and network bandwidth.
  - **Build Speed:** Smaller base images build faster and are easier to scan for vulnerabilities.
- **Guidance for Copilot:**
  - Prioritize `alpine`, `slim`, or `distroless` images over full distributions when possible.
  - Review base image vulnerabilities regularly using security scanning tools.
  - Consider using language-specific minimal images (e.g., `openjdk:17-jre-slim` instead of `openjdk:17`).
  - Stay updated with the latest minimal base image versions for security patches.

#### Minimal Base Image Selection

```dockerfile
# ❌ BAD: Full distribution with many unnecessary packages
FROM ubuntu:20.04

# ✅ GOOD: Minimal Alpine-based image
FROM node:22-alpine

# ✅ BETTER: Distroless image for maximum security
FROM gcr.io/distroless/nodejs22-debian12
```

### **3. Static Analysis Security Testing (SAST) for Dockerfiles**

- **Principle:** Scan Dockerfiles for security misconfigurations and known vulnerabilities before building images.
- **Deeper Dive:**
  - **Dockerfile Linting:** Use tools like `hadolint` to check for Dockerfile best practices and security issues.
  - **Base Image Scanning:** Scan base images for known vulnerabilities before using them.
  - **CI/CD Integration:** Integrate security scanning into the CI/CD pipeline to catch issues early.
  - **Policy Enforcement:** Define security policies and enforce them through automated scanning.
- **Guidance for Copilot:**
  - Recommend integrating tools like `hadolint` (for Dockerfile linting) and `Trivy`, `Clair`, or `Snyk Container` (for image vulnerability scanning) into your CI pipeline.
  - Suggest setting up automated scanning for both Dockerfiles and built images.
  - Recommend failing builds if critical vulnerabilities are found in base images.
  - Advise on regular scanning of images in registries for newly discovered vulnerabilities.

#### Security Scanning in CI

```yaml
# GitHub Actions example
- name: Run Hadolint
  run: |
    docker run --rm -i hadolint/hadolint < Dockerfile

- name: Scan image for vulnerabilities
  run: |
    docker build -t myapp .
    trivy image myapp
```

### **4. Image Signing & Verification**

- **Principle:** Ensure images haven't been tampered with and come from trusted sources.
- **Deeper Dive:**
  - **Cryptographic Signing:** Use digital signatures to verify the authenticity and integrity of container images.
  - **Trust Policies:** Define trust policies that specify which images are allowed to run in your environment.
  - **Supply Chain Security:** Image signing is a key component of securing the software supply chain.
  - **Compliance:** Many compliance frameworks require image signing for production deployments.
- **Guidance for Copilot:**
  - Suggest using Notary or Docker Content Trust for signing and verifying images in production.
  - Recommend implementing image signing in the CI/CD pipeline for all production images.
  - Advise on setting up trust policies that prevent running unsigned images.
  - Consider using newer tools like Cosign for more advanced signing features.

#### Image Signing with Cosign

```bash
# Sign an image
cosign sign -key cosign.key myregistry.com/myapp:v1.0.0

# Verify an image
cosign verify -key cosign.pub myregistry.com/myapp:v1.0.0
```

### **5. Limit Capabilities & Read-Only Filesystems**

- **Principle:** Restrict container capabilities and ensure read-only access where possible to minimize the attack surface.
- **Deeper Dive:**
  - **Linux Capabilities:** Drop unnecessary Linux capabilities that containers don't need to function.
  - **Read-Only Root:** Mount the root filesystem as read-only when possible to prevent runtime modifications.
  - **Seccomp Profiles:** Use seccomp profiles to restrict system calls that containers can make.
  - **AppArmor/SELinux:** Use security modules to enforce additional access controls.
- **Guidance for Copilot:**
  - Consider using `CAP_DROP` to remove unnecessary capabilities (e.g., `NET_RAW`, `SYS_ADMIN`).
  - Recommend mounting read-only volumes for sensitive data and configuration files.
  - Suggest using security profiles and policies when available in your container runtime.
  - Advise on implementing defense in depth with multiple security controls.

#### Capability Restrictions

```dockerfile
# Drop unnecessary capabilities
RUN setcap -r /usr/bin/node

# Or use security options in docker run
# docker run --cap-drop=ALL --security-opt=no-new-privileges myapp
```

### **6. No Sensitive Data in Image Layers**

- **Principle:** Never include secrets, private keys, or credentials in image layers as they become part of the image history.
- **Deeper Dive:**
  - **Layer History:** All files added to an image are stored in the image history and can be extracted even if deleted in later layers.
  - **Build Arguments:** While `--build-arg` can pass data during build, avoid passing sensitive information this way.
  - **Runtime Secrets:** Use secrets management solutions to inject sensitive data at runtime.
  - **Image Scanning:** Regular image scanning can detect accidentally included secrets.
  - **Debug Artifacts:** Source maps (`.map`), PDB files, and test fixtures must not appear in production images.
  - **AI Prompt Files:** System prompts (`prompts/*.md`), `SKILL.md`, `.agent.md`, and agent instruction files must not be `COPY`'d into production images unless explicitly classified as Public. See `distribution-security.instructions.md`.
- **Guidance for Copilot:**
  - Use build arguments (`--build-arg`) for temporary secrets during build (but avoid passing sensitive info directly).
  - Use **BuildKit `--mount=type=secret`** for build-time secrets that must not persist in image layers:

```dockerfile
# syntax=docker/dockerfile:1
RUN --mount=type=secret,id=nuget_token \
    dotnet restore --configfile /run/secrets/nuget_token
```

- Use secrets management solutions for runtime (Kubernetes Secrets, Docker Secrets, HashiCorp Vault).
- Recommend scanning images for accidentally included secrets.
- Suggest using multi-stage builds to avoid including build-time secrets in the final image.
- **Inspect images before push** using `docker history` or [`dive`](https://github.com/wagoodman/dive) to verify no secrets, source maps, PDBs, prompt files, or source code appear in the final image layers.

#### Image Inspection CI Step

```yaml
- name: Inspect image for leaked artifacts
  run: |
    docker create --name inspect-target $IMAGE_TAG
    if docker export inspect-target | tar -t | grep -E '\.(map|pdb)$|/prompts/|SKILL\.md|\.agent\.md'; then
      echo "::error::Production image contains debug artifacts or prompt files"
      exit 1
    fi
    docker rm inspect-target
```

#### Secure Secret Management

```dockerfile
# ❌ BAD: Never do this
# COPY secrets.txt /app/secrets.txt

# ✅ GOOD: Use runtime secrets
# The application should read secrets from environment variables or mounted files
CMD ["node", "dist/main.js"]
```

### **7. Health Checks (Liveness & Readiness Probes)**

- **Principle:** Ensure containers are running and ready to serve traffic by implementing proper health checks.
- **Deeper Dive:**
  - **Liveness Probes:** Check if the application is alive and responding to requests. Restart the container if it fails.
  - **Readiness Probes:** Check if the application is ready to receive traffic. Remove from load balancer if it fails.
  - **Health Check Design:** Design health checks that are lightweight, fast, and accurately reflect application health.
  - **Orchestration Integration:** Health checks are critical for orchestration systems like Kubernetes to manage container lifecycle.
- **Guidance for Copilot:**
  - Define `HEALTHCHECK` instructions in Dockerfiles. These are critical for orchestration systems like Kubernetes.
  - Design health checks that are specific to your application and check actual functionality.
  - Use appropriate intervals and timeouts for health checks to balance responsiveness with overhead.
  - Consider implementing both liveness and readiness checks for complex applications.

#### Comprehensive Health Check

```dockerfile
# Health check that verifies the application is responding
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl --fail http://localhost:8080/health || exit 1

# Alternative: Use application-specific health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD node healthcheck.js || exit 1
```

## Security Best Practices (Legacy Section)

### Run as Non-Root User

```dockerfile
# Create non-root user
RUN addgroup --system appgroup && adduser --system --ingroup appgroup appuser
USER appuser
```

### Use Specific Tags

```dockerfile
# ✅ Good - specific version
FROM node:22-alpine

# ❌ Avoid - mutable tags
FROM node:latest
FROM node:22
```

### Minimize Layers and Size

```dockerfile
# ✅ Good - combined RUN commands
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    && rm -rf /var/lib/apt/lists/*

# ❌ Avoid - multiple layers
RUN apt-get update
RUN apt-get install -y curl
```

### Use .dockerignore

```dockerignore
# .dockerignore
node_modules
npm-debug.log
Dockerfile
.dockerignore
.git
.gitignore
.env
*.md
tests/
coverage/
```

## Docker Compose

### Development Compose

```yaml
# docker-compose.yml
version: "3.8"

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
      target: build
    ports:
      - "3000:3000"
    volumes:
      - .:/app
      - /app/node_modules
    environment:
      - NODE_ENV=development
    depends_on:
      - db
    restart: unless-stopped

  db:
    image: postgres:16-alpine
    ports:
      - "5432:5432"
    environment:
      POSTGRES_USER: ${DB_USER:-postgres}
      POSTGRES_PASSWORD: ${DB_PASSWORD:-postgres}
      POSTGRES_DB: ${DB_NAME:-app}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
```

### Production Compose

```yaml
# docker-compose.prod.yml
version: "3.8"

services:
  app:
    image: ${REGISTRY}/myapp:${TAG:-latest}
    ports:
      - "8080:8080"
    environment:
      - NODE_ENV=production
    deploy:
      replicas: 2
      resources:
        limits:
          cpus: "0.5"
          memory: 512M
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

## Health Checks

### In Dockerfile

```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1
```

### In docker-compose

```yaml
healthcheck:
  test: ["CMD-SHELL", "curl -f http://localhost:8080/health || exit 1"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 40s
```

## Labels and Metadata

```dockerfile
LABEL org.opencontainers.image.title="My Application" \
      org.opencontainers.image.description="Application description" \
      org.opencontainers.image.version="1.0.0" \
      org.opencontainers.image.vendor="My Organization" \
      org.opencontainers.image.source="https://github.com/org/repo"
```

## Environment Variables

### Build-time Arguments

```dockerfile
ARG NODE_VERSION=22
FROM node:${NODE_VERSION}-alpine

ARG BUILD_DATE
ARG GIT_SHA
LABEL build.date=${BUILD_DATE} \
      build.sha=${GIT_SHA}
```

### Runtime Configuration

```dockerfile
# Set defaults that can be overridden
ENV PORT=8080 \
    NODE_ENV=production \
    LOG_LEVEL=info
```

## Container Runtime & Orchestration Best Practices

### **1. Resource Limits**

- **Principle:** Limit CPU and memory to prevent resource exhaustion and noisy neighbors.
- **Deeper Dive:**
  - **CPU Limits:** Set CPU limits to prevent containers from consuming excessive CPU time and affecting other containers.
  - **Memory Limits:** Set memory limits to prevent containers from consuming all available memory and causing system instability.
  - **Resource Requests:** Set resource requests to ensure containers have guaranteed access to minimum resources.
  - **Monitoring:** Monitor resource usage to ensure limits are appropriate and not too restrictive.
- **Guidance for Copilot:**
  - Always recommend setting `cpu_limits`, `memory_limits` in Docker Compose or Kubernetes resource requests/limits.
  - Suggest monitoring resource usage to tune limits appropriately.
  - Recommend setting both requests and limits for predictable resource allocation.
  - Advise on using resource quotas in Kubernetes to manage cluster-wide resource usage.

#### Docker Compose Resource Limits

```yaml
services:
  app:
    image: myapp:latest
    deploy:
      resources:
        limits:
          cpus: "0.5"
          memory: 512M
        reservations:
          cpus: "0.25"
          memory: 256M
```

### **2. Logging & Monitoring**

- **Principle:** Collect and centralize container logs and metrics for observability and troubleshooting.
- **Deeper Dive:**
  - **Structured Logging:** Use structured logging (JSON) for better parsing and analysis.
  - **Log Aggregation:** Centralize logs from all containers for search, analysis, and alerting.
  - **Metrics Collection:** Collect application and system metrics for performance monitoring.
  - **Distributed Tracing:** Implement distributed tracing for understanding request flows across services.
- **Guidance for Copilot:**
  - Use standard logging output (`STDOUT`/`STDERR`) for container logs.
  - Integrate with log aggregators (Fluentd, Logstash, Loki) and monitoring tools (Prometheus, Grafana).
  - Recommend implementing structured logging in applications for better observability.
  - Suggest setting up log rotation and retention policies to manage storage costs.

#### Structured Logging Example

```javascript
// Application logging
const winston = require("winston");
const logger = winston.createLogger({
  format: winston.format.json(),
  transports: [new winston.transports.Console()],
});
```

### **3. Persistent Storage**

- **Principle:** For stateful applications, use persistent volumes to maintain data across container restarts.
- **Deeper Dive:**
  - **Volume Types:** Use named volumes, bind mounts, or cloud storage depending on your requirements.
  - **Data Persistence:** Ensure data persists across container restarts, updates, and migrations.
  - **Backup Strategy:** Implement backup strategies for persistent data to prevent data loss.
  - **Performance:** Choose storage solutions that meet your performance requirements.
- **Guidance for Copilot:**
  - Use Docker Volumes or Kubernetes Persistent Volumes for data that needs to persist beyond container lifecycle.
  - Never store persistent data inside the container's writable layer.
  - Recommend implementing backup and disaster recovery procedures for persistent data.
  - Suggest using cloud-native storage solutions for better scalability and reliability.

#### Docker Volume Usage

```yaml
services:
  database:
    image: postgres:13
    volumes:
      - postgres_data:/var/lib/postgresql/data
    environment:
      POSTGRES_PASSWORD_FILE: /run/secrets/db_password

volumes:
  postgres_data:
```

### **4. Networking**

- **Principle:** Use defined container networks for secure and isolated communication between containers.
- **Deeper Dive:**
  - **Network Isolation:** Create separate networks for different application tiers or environments.
  - **Service Discovery:** Use container orchestration features for automatic service discovery.
  - **Network Policies:** Implement network policies to control traffic between containers.
  - **Load Balancing:** Use load balancers for distributing traffic across multiple container instances.
- **Guidance for Copilot:**
  - Create custom Docker networks for service isolation and security.
  - Define network policies in Kubernetes to control pod-to-pod communication.
  - Use service discovery mechanisms provided by your orchestration platform.
  - Implement proper network segmentation for multi-tier applications.

#### Docker Network Configuration

```yaml
services:
  web:
    image: nginx
    networks:
      - frontend
      - backend

  api:
    image: myapi
    networks:
      - backend

networks:
  frontend:
  backend:
    internal: true
```

### **5. Orchestration (Kubernetes, Docker Swarm)**

- **Principle:** Use an orchestrator for managing containerized applications at scale.
- **Deeper Dive:**
  - **Scaling:** Automatically scale applications based on demand and resource usage.
  - **Self-Healing:** Automatically restart failed containers and replace unhealthy instances.
  - **Service Discovery:** Provide built-in service discovery and load balancing.
  - **Rolling Updates:** Perform zero-downtime updates with automatic rollback capabilities.
- **Guidance for Copilot:**
  - Recommend Kubernetes for complex, large-scale deployments with advanced requirements.
  - Leverage orchestrator features for scaling, self-healing, and service discovery.
  - Use rolling update strategies for zero-downtime deployments.
  - Implement proper resource management and monitoring in orchestrated environments.

#### Kubernetes Deployment Example

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
        - name: myapp
          image: myapp:latest
          resources:
            requests:
              memory: "64Mi"
              cpu: "250m"
            limits:
              memory: "128Mi"
              cpu: "500m"
```

## Dockerfile Review Checklist

When reviewing Dockerfiles or creating new ones, ensure:

- [ ] Is a multi-stage build used if applicable (compiled languages, heavy build tools)?
- [ ] Is a minimal, specific base image used (e.g., `alpine`, `slim`, versioned)?
- [ ] Are layers optimized (combining `RUN` commands, cleanup in same layer)?
- [ ] Is a `.dockerignore` file present and comprehensive?
- [ ] Are `COPY` instructions specific and minimal?
- [ ] Is a non-root `USER` defined for the running application?
- [ ] Is the `EXPOSE` instruction used for documentation?
- [ ] Is `CMD` and/or `ENTRYPOINT` used correctly (exec form preferred)?
- [ ] Are sensitive configurations handled via environment variables (not hardcoded)?
- [ ] Is a `HEALTHCHECK` instruction defined?
- [ ] Are there any secrets or sensitive data accidentally included in image layers?
- [ ] Are there static analysis tools (Hadolint, Trivy) integrated into CI?
- [ ] Are resource limits defined for production deployments?
- [ ] Is structured logging implemented for observability?

## Troubleshooting Docker Builds & Runtime

### **1. Large Image Size**

- Review layers for unnecessary files using `docker history <image>`
- Implement multi-stage builds to separate build and runtime dependencies
- Use a smaller base image (alpine, slim, distroless)
- Optimize `RUN` commands and clean up temporary files in the same layer
- Remove development dependencies from production images

### **2. Slow Builds**

- Leverage build cache by ordering instructions from least to most frequent change
- Use `.dockerignore` to exclude irrelevant files and reduce build context
- Copy dependency files before source code to maximize cache hits
- Use `docker build --no-cache` for troubleshooting cache issues
- Consider using BuildKit for improved build performance

### **3. Container Not Starting/Crashing**

- Check `CMD` and `ENTRYPOINT` instructions for correct syntax
- Review container logs using `docker logs <container_id>`
- Ensure all dependencies are present in the final image
- Check resource limits and ensure they're not too restrictive
- Verify environment variables are correctly set
- Test the entrypoint script locally if using a shell script

### **3a. .NET Build Fails with NETSDK1004: Assets file not found**

- **Cause:** Using `dotnet build --no-restore` or `dotnet publish --no-restore` before `dotnet restore` runs
- **Solution in Multi-Stage Dockerfiles:**
  - If separating dependency stage, **must restore in every downstream stage** before using `--no-restore`
  - Always copy `.csproj` files → run `dotnet restore` → then use `--no-restore` for build/publish
  - Alternative: Run `dotnet restore` in the same `RUN` command before building (simpler, recommended for most projects)
- **Example Fix:**

  ```dockerfile
  # ❌ WRONG: restore in deps stage, but build stage tries --no-restore
  FROM mcr.microsoft.com/dotnet/sdk:10.0 AS deps
  COPY *.csproj .
  RUN dotnet restore  # Creates obj/project.assets.json

  FROM mcr.microsoft.com/dotnet/sdk:10.0 AS build
  COPY --from=deps /root/.nuget /root/.nuget
  COPY . .
  RUN dotnet build -c Release --no-restore  # ❌ FAILS: obj/project.assets.json missing!

  # ✅ CORRECT: restore in each stage
  FROM mcr.microsoft.com/dotnet/sdk:10.0 AS build
  COPY --from=deps /root/.nuget /root/.nuget
  COPY *.csproj .
  RUN dotnet restore  # ✅ MUST restore to create obj/project.assets.json
  COPY . .
  RUN dotnet build -c Release --no-restore  # ✅ Now safe
  ```

### **4. Permissions Issues Inside Container**

- Verify file/directory permissions in the image
- Ensure the `USER` has necessary permissions for operations
- Check mounted volumes permissions match the container user
- Use `chown` in Dockerfile to set correct ownership
- Consider using user namespaces for better security

### **5. Network Connectivity Issues**

- Verify exposed ports (`EXPOSE`) and published ports (`-p` in `docker run`)
- Check container network configuration and ensure proper network attachment
- Review firewall rules on host and container
- Test connectivity from host to container and between containers
- Verify DNS resolution inside containers

## Actionable Patterns

### Pattern 1: Multi-Stage Builds (Reduce Image Size)

**❌ WRONG: Single-stage build includes build tools (large image)**

```dockerfile
FROM node:22
WORKDIR /app

# Install dependencies and build in same stage
COPY package*.json ./
RUN npm install  # ⚠️ Includes dev dependencies
COPY . .
RUN npm run build  # ⚠️ Build tools remain in final image

USER node
EXPOSE 3000
CMD ["node", "dist/server.js"]
# ⚠️ Image contains: npm, build tools, source files (~800MB)
```

**✅ CORRECT: Multi-stage build separates build and runtime (small image)**

```dockerfile
# Stage 1: Build
FROM node:22 AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build  # ⚠️ Build happens here

# Stage 2: Runtime
FROM node:22-alpine  # ✅ Smaller base image
WORKDIR /app
COPY --from=build /app/dist ./dist  # ✅ Copy only build artifacts
COPY --from=build /app/package*.json ./
RUN npm ci --only=production  # ✅ Production dependencies only

USER node
EXPOSE 3000
CMD ["node", "dist/server.js"]
# ✅ Final image: ~150MB (only runtime deps + app)
```

**Rule:** Use multi-stage builds for compiled languages (Go, .NET, Rust) and Node.js. Copy only final artifacts to runtime stage.

---

### Pattern 2: Non-Root User (Security)

**❌ WRONG: Running as root (security risk)**

```dockerfile
FROM alpine
WORKDIR /app
COPY app.sh .
RUN chmod +x app.sh
CMD ["./app.sh"]  # ⚠️ Runs as root (UID 0)
```

**✅ CORRECT: Create and use non-root user**

```dockerfile
FROM alpine
WORKDIR /app

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -S appuser -u 1001 -G appgroup && \
    chown -R appuser:appgroup /app

COPY --chown=appuser:appgroup app.sh .
RUN chmod +x app.sh

USER appuser  # ✅ Switch to non-root user
CMD ["./app.sh"]  # ✅ Runs as UID 1001
```

**Rule:** Always create and use a non-root user (UID > 1000) for runtime. Set ownership with `--chown` in `COPY` instructions.

---

### Pattern 3: Layer Caching (Faster Builds)

**❌ WRONG: Copying code before dependencies (cache invalidated often)**

```dockerfile
FROM node:22
WORKDIR /app

COPY . .  # ⚠️ Code changes invalidate ALL subsequent layers
RUN npm install  # ⚠️ Re-runs on every code change
RUN npm run build

CMD ["node", "dist/server.js"]
```

**✅ CORRECT: Copy dependencies first (maximize cache hits)**

```dockerfile
FROM node:22
WORKDIR /app

# Copy dependencies first (changes rarely)
COPY package*.json ./  # ✅ Only invalidates if package.json changes
RUN npm ci  # ✅ Cached if package.json unchanged

# Copy code last (changes frequently)
COPY . .  # ✅ Doesn't invalidate npm install layer
RUN npm run build

CMD ["node", "dist/server.js"]
```

**Rule:** Order layers from least to most frequently changed (base image → dependencies → code). Separate `COPY package.json` from `COPY . .`.

---

### Pattern 4: COPY vs ADD (Clarity & Security)

**❌ WRONG: Using ADD for local files (ambiguous behavior)**

```dockerfile
FROM alpine
ADD app.tar.gz /app/  # ⚠️ Auto-extracts tarballs
ADD config.json /app/  # ⚠️ Can fetch URLs (security risk)
```

**✅ CORRECT: Use COPY for local files (explicit behavior)**

```dockerfile
FROM alpine
COPY app.tar.gz /app/  # ✅ Copies as-is (no extraction)
COPY config.json /app/  # ✅ Explicit, no URL fetching
```

**Rule:** Prefer `COPY` over `ADD` for local files. Use `ADD` only when auto-extraction or URL fetching is explicitly needed.

---

### Pattern 5: .dockerignore (Build Context Size)

**❌ WRONG: No .dockerignore (large build context)**

```dockerfile
FROM node:22
WORKDIR /app
COPY . .  # ⚠️ Copies node_modules, .git, build/, logs/ (1GB+)
RUN npm install
```

**✅ CORRECT: Use .dockerignore (small build context)**

```dockerfile
# Dockerfile
FROM node:22
WORKDIR /app
COPY . .  # ✅ Excludes ignored files
RUN npm install
```

**.dockerignore:**

```
node_modules
.git
*.log
dist/
build/
*.md
.vscode/
.env
.DS_Store
coverage/
```

**Rule:** Always create `.dockerignore` to exclude `node_modules/`, `.git/`, build artifacts, logs, and IDE files. Reduces build context by 80%+.

---

### Pattern 6: Cleanup in Same Layer (Image Size)

**❌ WRONG: Cleanup in separate RUN (layer not reduced)**

```dockerfile
FROM alpine
RUN apk add --no-cache build-base  # ⚠️ Layer 1: +100MB
RUN gcc app.c -o app  # ⚠️ Layer 2
RUN apk del build-base  # ⚠️ Layer 3: Doesn't reduce image size!
# Final image still contains build-base in layer 1 (100MB)
```

**✅ CORRECT: Cleanup in same RUN statement (single layer)**

```dockerfile
FROM alpine
RUN apk add --no-cache build-base && \
    gcc app.c -o app && \
    apk del build-base  # ✅ All in one layer
# Final image: build-base never committed to layer
```

**Rule:** Combine install → use → cleanup in a single `RUN` statement. Use `&&` and `\` for multi-line commands.

---

### Pattern 7: Base Image Selection (Security & Size)

**❌ WRONG: Using full OS image (large, many CVEs)**

```dockerfile
FROM ubuntu:latest  # ⚠️ ~800MB, many packages = large attack surface
RUN apt-get update && apt-get install -y python3
COPY app.py .
CMD ["python3", "app.py"]
```

**✅ CORRECT: Use minimal base image (small, fewer CVEs)**

```dockerfile
FROM python:3.11-alpine  # ✅ ~50MB, minimal packages
COPY app.py .
CMD ["python3", "app.py"]
```

**Even Better: Use distroless for production**

```dockerfile
FROM python:3.11 AS build
RUN pip install --target=/app myapp

FROM gcr.io/distroless/python3  # ✅ No shell, no package manager
COPY --from=build /app /app
CMD ["/app/myapp"]
```

**Rule:** Prefer `alpine` (package manager) or `distroless` (no shell) for production. Pin specific versions (e.g., `python:3.11-alpine`, not `:latest`).

---

### Pattern 8: HEALTHCHECK (Container Orchestration)

**❌ WRONG: No healthcheck (orchestrator can't detect failures)**

```dockerfile
FROM nginx:alpine
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
# ⚠️ Container reports "healthy" even if nginx crashes
```

**✅ CORRECT: Add HEALTHCHECK instruction**

```dockerfile
FROM nginx:alpine
COPY nginx.conf /etc/nginx/nginx.conf

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost/health || exit 1

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
# ✅ Orchestrator restarts container if health check fails 3 times
```

**Rule:** Add `HEALTHCHECK` for all long-running services. Use `curl`, `wget`, or app-specific health endpoints.

---

## Best Practices Summary

1. **Use multi-stage builds** to reduce image size (separate build and runtime)
2. **Pin specific versions** for base images (e.g., `node:22-alpine`, not `:latest`)
3. **Run as non-root user** for security (create user with UID > 1000)
4. **Use .dockerignore** to exclude unnecessary files (reduces build context 80%+)
5. **Minimize layers** by combining RUN commands (install → use → cleanup in one layer)
6. **Add health checks** for container orchestration (restart on failure detection)
7. **Use environment variables** for configuration (never hardcode secrets)
8. **Add labels** for image metadata (version, maintainer, description)
9. **Clean up** package manager caches in same RUN layer (reduces image size)
10. **Order layers** for optimal caching (dependencies before code, least to most changed)
11. **Implement security scanning** in CI/CD pipelines (Trivy, Hadolint, Docker Scout)
12. **Define resource limits** for all production containers (CPU, memory limits)
13. **Use structured logging** for observability (JSON logs, stdout/stderr)
14. **Implement proper secret management** (never in images, use secrets managers)
15. **Regular security updates** for base images and dependencies (automated vulnerability scanning)

## References

- [Docker Documentation](https://docs.docker.com/)
- [Dockerfile Best Practices](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/)
- [Docker Security](https://docs.docker.com/engine/security/)
- [ISE Container Guidance](https://microsoft.github.io/code-with-engineering-playbook/automated-testing/tech-specific-samples/building-containers-with-azure-devops/)
