## HPK - High Performance Kubernetes
HPC-style batch jobs in Kubernetes. Get it?

There is a need to leverage cloud-native technologies within the HPC eco-system.

Both commands assume there's an established `kubeconfig` on the client machine.

### Krun
Execute a command remotely, block, and immediately return results to the CLI.

*Example* `krun -numnodes 3 hostname`

Expected result: 3 hostnames (`-numnodes 3` means run 3 different instances)

### Kbatch
Submit a bash script for non-blocking execution across ephemeral nodes.

*Example* `kbatch test.sh`

Expected result: Nothing yet pipes back to the client shell. It will run `test.sh` as a k8s job with test.sh in the container home directory.

### Kapi / Kdash
API server and dashboard for HPK jobs. Dashboard also has a config editor.

### Roadmap:
- Client in go, mirror srun / sbatch flags
- Jobs mount same scratch (not just emptydir)
- Home directories / UID/GID
- kjobapi server to track allocations and job status
