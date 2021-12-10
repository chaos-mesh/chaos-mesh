# Pipeline Controller

In early implementation of controllers in v2.0, all common controllers are registered respectively 
and may be executed by different order, which result in bugs like 
[#2449](https://github.com/chaos-mesh/chaos-mesh/issues/2449). So we introduce a pipeline controller 
to register them and execute them in a fixed order.
