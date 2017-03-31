#!/bin/bash

# 1.  Installer creates namespace of its choice
# 2.  Installer creates `api` service with a fixed selector (app: api) with service serving cert annotation
# 3.  Installer creates `apiserver` service account
# 4.  Installer harvests serving cert
# 5.  Installer harvests SA token
# 6.  Installer creates SA kubeconfig
# 7.  Installer binds “known” roles to SA user
# 8.  Installer provides the following information to all templates.
# 	1. etcd location
# 	2. etcd ca location
# 	3. etcd write cert/key path
# 	4. serving cert/key path
# 	5. SA kubeconfig path
# 	6. namespace name
# 9.  Installer processes and creates the pre-create template - this is the template that provides things like clusterroles, clusterrolebindings, random weird stuff.
# 10. Installer processes the static pod template - this is a template that ONLY creates a static pod.
# 11. Installer takes the pod and adds it to the node pod manifest folder
# 12. Installer waits until all the pod’s containers are ready.
# 13. Installer creates the APIService object and registers it

targetNamespace=brokersdk
#apiserverConfigDir=/home/bparees/git/gocode/src/github.com/openshift/origin/examples/sample-app/openshift.local.config/brokersdk
#masterConfigDir=/home/bparees/git/gocode/src/github.com/openshift/origin/examples/sample-app/openshift.local.config/master
#mkdir -p ${apiserverConfigDir} || true
#nodeManifestDir=openshift.local.config/node-deads-dev-01/static-pods


# 1.  Installer creates namespace of its choice
oc new-project ${targetNamespace}

# 2.  Installer creates `api` service with a fixed selector (app: api) with service serving cert annotation
#oc -n ${targetNamespace} create service clusterip brokersdk --tcp=443:443
oc create -f resources/service.yaml

#oc -n ${targetNamespace} annotate svc/brokersdk service.alpha.openshift.io/serving-cert-secret-name=brokersdk-serving-cert
#until oc -n ${targetNamespace} get secrets/brokersdk-serving-cert; do
#	echo "waiting for oc -n ${targetNamespace} get secrets/brokersdk-serving-cert"
#	sleep 1
#done

# 3.  Installer creates `brokersdk` service account
#oc  -n ${targetNamespace} create sa brokersdk
oc create -f resources/sa.yaml

#until oc -n ${targetNamespace} sa get-token brokersdk; do
#	echo "waiting for oc -n ${targetNamespace} get secrets/brokersdk-serving-cert"
#	sleep 1
#done

# 4.  Installer harvests serving cert
#oc -n ${targetNamespace} extract secret/brokersdk-serving-cert --to=${apiserverConfigDir}
#mv ${apiserverConfigDir}/tls.crt ${apiserverConfigDir}/serving.crt
#mv ${apiserverConfigDir}/tls.key ${apiserverConfigDir}/serving.key

# 5.  Installer harvests SA token
#saToken=$(oc -n ${targetNamespace} sa get-token brokersdk)

# 6.  Installer creates SA kubeconfig
# TODO do this a LOT better
# start with admin.kubeconfig
#cp ${masterConfigDir}/admin.kubeconfig ${apiserverConfigDir}/kubeconfig
# remove all users
#oc --config=${apiserverConfigDir}/kubeconfig config unset users
# set the service account token
#configContext=$(oc --config=${apiserverConfigDir}/kubeconfig config current-context)
#oc --config=${apiserverConfigDir}/kubeconfig config set-credentials serviceaccount --token=${saToken}
#oc --config=${apiserverConfigDir}/kubeconfig config set-context ${configContext} --user=serviceaccount

# 7.  Installer binds “known” roles to SA user
# TODO remove this bit once we bootstrap these roles
oc create -f resources/roles.yaml || true

oadm policy add-cluster-role-to-user system:auth-delegator -n ${targetNamespace} -z brokersdk
oc create policybinding kube-system -n kube-system
oadm policy add-role-to-user extension-apiserver-authentication-reader -n kube-system --role-namespace=kube-system system:serviceaccount:${targetNamespace}:brokersdk

# allow us to run the broker pods as root.
oadm policy add-scc-to-user anyuid -z brokersdk


# 8.  Installer provides the following information to all templates.
#cp ${masterConfigDir}/ca.crt ${apiserverConfigDir}/etcd-ca.crt
#cp ${masterConfigDir}/ca-bundle.crt ${apiserverConfigDir}/client-ca.crt
#cp ${masterConfigDir}/master.etcd-client.crt ${apiserverConfigDir}/etcd-write.crt
#cp ${masterConfigDir}/master.etcd-client.key ${apiserverConfigDir}/etcd-write.key
#templateArgs="CLIENT_CA=/etcd/apiserver-config/client-ca.crt"
#templateArgs="${templateArgs} SERVING_CRT=/etcd/apiserver-config/serving.crt"
#templateArgs="${templateArgs} SERVING_KEY=/etcd/apiserver-config/serving.key"
#templateArgs="${templateArgs} KUBECONFIG=/etcd/apiserver-config/kubeconfig"
#templateArgs="${templateArgs} NAMESPACE=${targetNamespace}"
#templateArgs="${templateArgs} CONFIG_DIR=${apiserverConfigDir}"
#templateArgs="${templateArgs} CONFIG_DIR_MOUNT=/etcd/apiserver-config"

# 9.  Installer processes and creates the pre-create template - this is the template that provides things like clusterroles, clusterrolebindings, random weird stuff.
# nothing for wardle

# 10. Installer processes the static pod template - this is a template that ONLY creates a static pod.
#oc process -f artifacts/podtemplate.yaml ${templateArgs} | jq .items[0] > ${nodeManifestDir}/${targetNamespace}.yaml
oc create -f resources/rc.yaml

# curl -k https://172.30.92.139:8443/apis/generic.broker.k8s.io/v1alpha1/serviceinstances
# curl -k https://172.30.92.139:8443/broker/my.broker.io/v2/catalog
# curl -k -X PUT -H 'X-Broker-API-Version: 2.9' -H  'content-type: application/json' https://172.30.4.53:8443/broker/my.broker.io/v2/service_instances/1234
# curl -X PUT -H 'X-Broker-API-Version: 2.9' -H 'Content-Type: application/json' -d @provision.json -k  https://172.30.71.39:8443/broker/my.broker.io/v2/service_instances/1234
# curl -X DELETE -H 'X-Broker-API-Version: 2.9' -H 'Content-Type: application/json' -k  https://172.30.71.39:8443/broker/my.broker.io/v2/service_instances/1234?accepts_incomplete=true

