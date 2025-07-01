# Agents Documentation Overview

This document consolidates the content of **all** markdown (`*.md`) files in the repository. Use it as a single entry-point to understand the purpose, structure, and usage of the project as well as the assorted examples, development guides, and supporting material that accompany it.

_Last generated automatically by the documentation agent on 2025-07-01._

---

## Table of Contents

1. Root Documentation
   * [README.md](README.md)
   * [USAGE.md](USAGE.md)
   * [CMAKE_FILE_API.md](CMAKE_FILE_API.md)
   * [GIT_HOOKS.md](GIT_HOOKS.md)
2. Development Guides
   * [AGENTS.md](AGENTS.md) _(this file)_
   * [CLAUDE.md](CLAUDE.md)
3. Examples
   * [examples/README.md](examples/README.md)
   * [examples/simple_hello/README.md](examples/simple_hello/README.md)
   * [examples/libzmq/README.md](examples/libzmq/README.md)
   * [examples/cmake_directive/README.md](examples/cmake_directive/README.md)
4. Test Data
   * [testdata/external_cmake_project/README.md](testdata/external_cmake_project/README.md)
5. Micro-Agents
   * [.openhands/microagents/repo.md](.openhands/microagents/repo.md)

---

## Summaries

### README.md  
Bazel Gazelle _foreign-cc_ plugin overview. Provides a high-level description, CI badge placeholder, available example projects, current status and feature roadmap, prerequisites (Bazel, Go, C++ toolchain), and basic build/development commands.

### USAGE.md  
Hands-on usage guide. Covers quick-start steps, prerequisites (Bazel+bzlmod, Go, CMake), two main usage patterns (local CMake projects vs. external projects via `gazelle:cmake` directive), detailed configuration directive reference, explanation of internal processing pipeline, supported CMake constructs, example walkthroughs, current capabilities, work-in-progress list, limitations, and troubleshooting tips.

### CMAKE_FILE_API.md  
Technical deep-dive into the CMake File API integration that has replaced legacy regex parsing. Describes core Go components (`cmake_api.go`, `generate.go`, `resolve.go`), key features (accurate target parsing, dependency resolution, include handling, sub-directory support, error handling), how to invoke the integration, testing strategy, benefits vs. regex, limitations, and planned enhancements (caching, incremental updates).

### GIT_HOOKS.md  
Instructions for setting up the repository's _pre-commit_ Git hook which runs `buildifier` to auto-format Bazel files. Covers installation, usage, bypassing/manual execution, and hook behaviour details.

### AGENTS.md & CLAUDE.md  
Both files currently contain an identical developer guide that introduces the project, outlines core functionality, project directory structure, key Go source files, Bazel module dependencies, rule generation patterns, development commands, Gazelle language interface expectations, supported CMake directives, rule generation strategy, implementation status checklist, and coding conventions.

### examples/README.md  
Explains the separate _examples_ Bazel module, lists available example projects (`simple_hello`, `libzmq`), and shows how to build examples and regenerate BUILD files with Gazelle. Also provides guidance for adding new examples.

### examples/simple_hello/README.md  
Walk-through for a minimal "Hello World" C++ example: file inventory, build and run commands, and instructions for regenerating BUILD.bazel via Gazelle.

### examples/libzmq/README.md  
Demonstrates integration with the real-world ZeroMQ (libzmq) library. Includes an overview, build and Gazelle commands, file list, and expected outcomes after rule generation (ability to build libzmq from Bazel).

### examples/cmake_directive/README.md  
End-to-end scenario showcasing the `gazelle:cmake` directive for an external repository. Details the MODULE.bazel `http_archive` definition, minimal BUILD.bazel stub containing the directive, Gazelle invocation, CMake sample project, and the resulting Bazel rules produced by Gazelle (library and binary targets with includes, deps, visibility).

### testdata/external_cmake_project/README.md  
Short note explaining that the directory is used to test `gazelle:cmake` handling for external repositories.

### .openhands/microagents/repo.md  
Internal micro-agent specification for this repository: concise description of the plugin's intent, repository governance rules (e.g., ensure Bazel targets build, avoid scope creep, use feature branches, never push to main), and tooling assumptions (Bazel and CMake pre-installed).

---

## How to Regenerate This File

Execute the documentation agent from the project root:

```bash
bazel run //:generate_agents_md
```

(If such a rule does not yet exist, run the script or command that reads `**/*.md` files and writes `AGENTS.md` with updated summaries.)

---

_This consolidated documentation is intended to provide contributors a single, up-to-date reference while keeping the original markdown files focused and maintainable._