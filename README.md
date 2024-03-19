# Naavik

Naavik assists in unification of configurations across services deployed with in Service Mesh. 

It acts upon a custom resource called TrafficConfig and transforms them into service mesh understandable resources.

When dependencies are expressed between clients and a service, Naavik can translate the configurations in 
TrafficConfig to all the clients of the service deployed across multiple clusters over service mesh

# High level Architecture Diagram

![High level Architecture Diagram](./docs/images/architecture_diagram.png)

# How to run Naavik locally
- Refer to [Development Guide](./docs/DEVELOPER.MD).
- More details on internals of Naavik can be found [here](./docs/)

