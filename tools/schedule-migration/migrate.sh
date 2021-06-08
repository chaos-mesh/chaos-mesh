NAMESPACES=$(kubectl get namespace | sed '1d' | awk '{print $1}')
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
CRDS="awschaos
dnschaos
gcpchaos
iochaos
jvmchaos
kernelchaos
networkchaos
podchaos
stresschaos
timechaos
"
cnt=0
build () {
    cd $SCRIPT_DIR
    go build main.go
}

update_yaml () {
    local yaml=$1
    ./main $yaml $yaml
}

reapply_crd () {
    kubectl delete -f ../../manifests/crd.yaml
    kubectl apply -f ../../manifests/crd.yaml
}

handle_namespace () {
    local namespace=$1
    for kind in $CRDS
    do
        echo "  searching resources $kind"
        resources=$(kubectl get $kind -n $namespace | sed '1d' | awk '{print $1}')
        for resource in $resources
        do
            echo "      getting $resource"
            kubectl get $kind $resource -n $namespace -o yaml > $cnt.yaml
            update_yaml $cnt.yaml
            let cnt++
        done
    done
}

build

for ns in $NAMESPACES
do
    echo "searching namespace $ns"
    handle_namespace $ns
done

reapply_crd

for (( id=0; id<$cnt; id++ ))
do
    kubectl apply -f $id.yaml
done
