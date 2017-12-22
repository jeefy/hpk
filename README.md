## HPK - High Performance Kubernetes
HPC-style batch jobs in Kubernetes. Get it?

There is a need to leverage cloud-native technologies within the HPC eco-system.

Both commands assume there's an established `kubeconfig` on the client machine.

### Krun
Execute a command remotely, block, and immediately return results to the CLI.

*Example* `krun -n 3 hostname`

Expected result: 3 hostnames (`-n 3` means run on 3 different instances)

### Kbatch
Submit a bash script for non-blocking execution across ephemeral nodes.

*Example* `kbatch test.sh`

Expected result: Nothing yet pipes back to the client shell. It will run `test.sh` as a k8s job with test.sh in the container home directory.

### Roadmap:
- Client in go, mirror srun / sbatch flags
- Jobs mount same scratch (not just emptydir)
- Home directories / UID/GID
- Pipe job start / job end to mongodb
- kjobapi server to track allocations and job status
