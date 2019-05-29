#!/bin/bash

case $1 in
  "-h" )
    echo "
    Usage:

      1) upgrade-interactive bind {repository}/{chart}
        Helm2 does not return repository for the installed chart. In some cases tool can find same chart name in multiple repos - this chard is marked as deprecated. In such case we can manually bind chart to selected repository.
        
        Example:
          helm upgrade-interactive bind jetstack/cert-manager

        In this case 'helm search cert-manager' tells that chart is situated in 'jetstack/cert-manager' and 'stable/cert-manager'. Now versions will be always checked against 'jetstack' repository only.

      2) upgrade-interactive ...
        We can pass usual flags we pass to 'helm upgrade'

        Example:
          helm upgrade-interactive --force --recreate-pods --atomic
    "
    ;;

  "bind" )
    $HELM_PLUGIN_DIR/bin/parser -bind $2
    ;;

  *)
    input=$(mktemp /tmp/helm-upgrade-interactive.XXXXXX)
    output=$(mktemp /tmp/helm-upgrade-interactive.XXXXXX)


    $HELM_BIN list --output json > $input

    $HELM_PLUGIN_DIR/bin/parser -input $input -output $output

    while IFS=" " read -r release chart version
    do
      cmd="$HELM_BIN upgrade $release $chart --version $version --reuse-values $*"
      echo ">>> $cmd"
      $cmd
    done < $output

    rm -f $input
    rm -f $output
    ;;
esac