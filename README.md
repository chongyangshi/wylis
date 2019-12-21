# Wylis

Wylis is a [Kubernetes]([https://kubernetes.io/](https://kubernetes.io/)) [DaemonSet]([https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/)) for generating constant [RPC]([https://en.wikipedia.org/wiki/Remote_procedure_call](https://en.wikipedia.org/wiki/Remote_procedure_call)) traffic between all nodes within a cluster, one which runs across multiple private networks with no private routing in between. It achieves two purposes:

* **Hold the door**: On a best-effort basis, prevents encapsulated ([IP-in-IP](https://en.wikipedia.org/wiki/IP_in_IP)) traffic tunnelled between private networks -- such as those used by [Calico](https://www.projectcalico.org/) -- from being dropped under a sporadic failure condition by [WireGuard](http://wireguard.com/) before they can be routed to the right destination. More details linked in the next section.
* **Measure RPC latency**: Measures the success and failure rate, as well as latencies of successful RPCs between all pairs of nodes. These are exported as [Prometheus](https://prometheus.io/) metrics for scrapping.

## Rationale

By sending periodic keep-alive packets in the form of RPC HTTP requests, Wylis prevents WireGuard instances tunnelling encapsulated IP-in-IP packets from being dropped due to an unknown cold-path packet drop issue.

Please see my [blog post](https://blog.scy.email/running-a-low-cost-distributed-kubernetes-cluster-on-bare-metal-with-wireguard.html) for more detail.

## Installation

Wylis acts as a Kubernetes API client within the cluster to periodically update its knowledge of other Wylis pods running on all other nodes within the cluster. It intentionally does not use any service proxy, as all pairs of nodes should have encapsulated traffic passed through in both directions.

The default [`wylis.yaml`](https://github.com/icydoge/wylis/tree/master/wylis.yaml) contains a ready to use configuration, with RBAC support, using the Wylis image from Docker Hub:

    git clone https://github.com/icydoge/wylis.git
    cd wylis
    less wylis.yaml # Never apply any Kubernetes manifest from the internet without a careful inspection
    kubectl apply -f wylis.yaml

## Development

If you are interested enough in running Wylis, I host my development builds of Wylis in my private Docker repository (`172.16.16.3:2443`), therefore it is probably easier to fork this repository and edit [`Dockerfile`](https://github.com/icydoge/wylis/tree/master/Dockerfile) to point it towards your own repository. 

You can then edit The default [`wylis.yaml`](https://github.com/icydoge/wylis/tree/master/wylis.yaml) to use your own build.

## Health Warning

* Wylis is not suitable for commercial, production use for obvious reasons. It is unlikely under a production-grade budget environment Wylis' use-case is needed.
* Because every pair of nodes will have periodic traffic in both directions (at a defined interval you can configure), the amount of periodic traffic in the cluster will be [O(n²)](https://en.wikipedia.org/wiki/Big_O_notation) as the number of nodes in your cluster grow. While the blank HTTP RPC requests used by Wylis are very cheap, you should still be wary if you have a really large cluster. This is also the reason I decided not to implement Wylis as a more idiomatic `watch` controller.
* Even though Wylis should effectively hold the door open, it is still recommend that you deploy a service proxy with built-in retries for time-outs, such as [Envoy](https://www.envoyproxy.io/), to improve the reliability of your tunnelled RPCs through retries.

## Where does the name of the project come from?

Without infringing on any applicable trademark rights, Googling "Wylis" should give you a good answer :) 
