# Reference Files and Documentation

The following documents are maintained and updated throughout the project. They are intended to help reduce implementation time and cost by serving as references.
All documents must always reflect the current state of the code.
- When requirements, behavior, or UI details are clarified during implementation, reflect them in the appropriate reference document immediately.

- `SPEC.md`: Defines the contract and specifications of the project
- `docs/`: Project-related documentation
    - `docs/works/`: Project work documents
    - `docs/lessons/`: Records lessons learned, improvements, and usage of external libraries during the project

# Work Documentation Rules
- For every task execution, use the `working-docs` skill to record work logs and decisions so they can be referenced in future tasks

# Task Completion
- Continue writing or modifying code without stopping until the objective is achieved
- Once code is written, perform build; if it fails, fix the issues

# Failure Handling and Retry Policy
- Do not repeat failures using the same approach. If the same issue persists, change the approach before retrying

# Code Quality Management
- Code must function as living documentation, avoiding excessive complexity and maintaining readability
- Perform refactoring after every task:
  - Use abbreviations for identifiers when they do not cause ambiguity (e.g., use `ctx` instead of `context`)
  - Avoid excessive nested loops
  - Each function must follow the single responsibility principle, but avoid over-fragmentation that increases structural complexity
- Avoid creating too many files in the project root directory:
  - If a component is loosely coupled with the main crate and highly reusable, separate it into its own crate
  - If files are separated but share a highly similar context, merge them into a single module or file

# Preferred Language
- Unless otherwise specified, all documentation and explanations, including work documents, should be written in Korean

# Read-only docs
Do not modify content below directories without intentional request by user:
- `docs/refs` includes reference documents for writing codes.
- `docs/reqs` includes long request documents for this project.