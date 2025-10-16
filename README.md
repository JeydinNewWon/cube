# Cube Orchestrator

A fun little project creating a small scale orchestrator, simulating the behaviour of Kubernetes and Google's Borg for educational purposes. Adapted from *"Build your own orchestrator in Go"* book.

## Modules and roles 

- `Scheduler` is an interface designed for implementing different scheduling algorithms. It decides what worker to allocate an incoming task toward. 
- `Task` is a struct representing a single unit of 'work' that needs to be scheduled. It features a list of properties analogous to a Docker container. It features both a required amount of memory, disk and cpu resourcing, which is metadata given to the Scheduler to allocate.
    - `TaskEvent` is responsible for describing a change in a `Tasks` status. It features the state that it next wants to be transitioned towards. 
- `Worker` is a struct representing a machine. Specifically, it represents a machine that executes and works with logical workloads (i.e. `Tasks`). 
- A `Node` represents the physical aspect of a machine. It features additional things like CPU, memory and disk resourcing requirements to represent the machine(s) themselves.
- A `Manager` actually handles the enqueuing of `Tasks` onto `Workers`. Whilst `Scheduler` is responsible for picking the next `Worker`, the `Manager` executes that and stores the metadata for that `Task` in itself. It contains internal information of the `Tasks`, `Workers` in the system and features additional convenient fields that handle mapping of Tasks to Workers, and vice versa.
