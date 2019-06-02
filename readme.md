# helm-interactive-upgrade

> Interactively upgrades selected chart to their latest versions

***This software is for internal purposes only, use it at your own risk***

MACOS ONLY: for now, this plugin compiled for mac-compatible devices only. In case of other architecture you shoud manually compile `bin/parser` binary using the following command:
```
# cd upgrade-interactive-parser && go build -o ../bin/parser
```

## Install
```
# helm plugin install https://github.com/afoninsky/helm-upgrade-interactive
# helm upgrade-interactive
```

## Usage

```
# helm upgrade-interactive -h
```

### # upgrade-interactive ...
We can pass usual flags we pass to 'helm upgrade'.
```
helm upgrade-interactive --force --recreate-pods --atomic
```

### # upgrade-interactive bind {repository}/{chart}
Helm2 does not return repository for the installed chart. In some cases tool can find same chart name in multiple repos - this chard is marked as deprecated. In such case we can manually bind chart to selected repository.

```
helm upgrade-interactive bind jetstack/cert-manager
```
In this case 'helm search cert-manager' tells that chart is situated in 'jetstack/cert-manager' and 'stable/cert-manager'. Now versions will be always checked against 'jetstack' repository only.