TODO - delete this file before final merge to main


1. Fix the pointers in make files to use the proper License text
2. Merge the old wire way to start the controller with the new
kubebuilder scaffolding.  Requires: 
   -  merging cmd/main.go with cmd/msm-ns/main.go
	- Just delete cmd/msm-ns
   -  merging internal/controller/streamdata_controller.go with internal/core/core.go and internal/core/wire.go
	- For streammapper add structure to "type StreamdataReconciler struct"
	- For node mapper the controller can easily watch nodes.  Need to
	  decide what design we want.  
		1. msm-nc watch all nodes and take an action on each
		2. Add a node field to the CRD and let the CP indicate the node
		3. msm-nc include the node in the status field with a state
3. helm updates
   -- Update the msm-deployments helm charts.  CRDs - will need to complete 
before NC can be deployed. maybe cp as well.  
   -- Create helm charts in msm-nc for the CRDs.  Discussion point the config
directory has base yamls for many things we need RBAC, Deployment, CRDs.  These
should form the basis of the helm charts,  but ideally some degree of generation is sued so when things change its not manual updates.  
4. Update the CP to CRUD the CRDs. Incremental
step would be to just use a string based template in go with the normal
kube client. Better way would be import proper client libs
5. Add reconcilation loop logic.
5A. Hook the current logic in pkg/stream-mapper/stream_mapper.go into the 
reconcile loop.
6. Update internal/controller/suite_test.go
7. Update .jenkins as needed
8. Documentation 
9. Testing ???
10. Consolidate logging or just use zap logger?
11. Bring the logging and other default configuration into the manager
12. Do we care that many things are just called "manager" vs "msm-network-controller".  Changes required in Dockerfile, Makefile and some of the config/manifests


