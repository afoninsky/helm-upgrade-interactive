***Данный софт находится в альфа-версии и выложен для получения раннего фидбека. Вы используете его на свой страх и риск.***


## Install
```
# helm plugin install https://github.com/afoninsky/helm-upgrade-interactive
# helm upgrade-interactive
```

### Usage
```
# helm upgrade-interactive -h
  
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
```