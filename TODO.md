- [x] Setup local k8s environment
    - [x] Choose the implementation
      Choose `minikube`
    - [x] Test it with a test deployment
- [-] Choose the Go implementation to copy/use
    - [x] Make sure to credit the author
    - [ ] Read it and comment it
- [ ] Break it down "atomic" components
  The different deployments that will work together to resolve the issue
- [ ] Run it and test it without optimization
    - [ ] Make sure programs don't crash
    - [ ] Make sure that autoscaling work
      Do we want autoscaling in v1?
    - [ ] Make sure that networking work
    - [ ] Make sure it returns the correct result
- [ ] k8sify the program
    - [ ] Add health endpoint
    - [ ] Add readiness endpoint
    - [ ] Add liveness endpoint
- [ ] Add monitoring and logging
    - [ ] How long does each request take to make
    - [ ] How long does each pod take to start
    - [ ] See what else needs monitoring
- [ ] Start optimizing!
