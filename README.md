# Relay

Relay is a platform for managing personal and team tasks with a focus on clear
ownership, collaboration, and cross-team coordination.

## Overview

Managing tasks across individuals and teams often leads to:

- Lack of visibility into ownership
- Poor coordination between teams
- Difficulty tracking progress in real time

Relay addresses these challenges by providing a unified system for task
management and collaboration.

## Development

### Prerequisites:

- docker
- go (v1.26+)
- git

Use `rcli`, a cli tool only for this repository.

You need to install `rcli` first.
  - Copy this repository
  - Run `go install ./rcli` from the root directory

Now to setup the development environment:


```bash
rcli project setup
```

### Frontend Setup

To set up the frontend, run the following commands:

```bash
cd ./services/web-v2
pnpm install
pnpm start
```

### Backend Setup

To set up the backend, run the following commands:

```bash
docker compose up --build postgres
rcli project setup db # if you have atlas installed then add a -a flag
```

To start working with the backend, you will need to install these things:
 - buf
 - atlas (database migration tool)

## Features

- **Task Management**  
  Create, update, and track tasks with clear ownership and status.

- **Team Collaboration**  
  Assign tasks to team members and collaborate efficiently within teams.

- **Cross-Team Coordination**  
  Assign and manage tasks across multiple teams with shared visibility.

- **Personal Task Tracking**  
  Manage individual tasks alongside team responsibilities in one place.

- **Progress Tracking**  
  Monitor task status and ensure accountability across workflows.

## Use Cases

- Managing team-based projects  
- Coordinating work across multiple teams  
- Tracking personal and shared tasks in a single system  

## Future Improvements

- Real-time updates and notifications  
- Task dependencies and workflow automation  
- Role-based access control (RBAC) and Attribute Based Access Control (ABAC)  
- Analytics and reporting  

## License

This project is licensed under the MIT License.


## Common Command

To learn about more commands, go to ./rcli/README.md
