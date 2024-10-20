default:
    echo 'Hello, world!'

[group('docs')]
docs-serve:
    cd mdbook && mdbook serve -n 0.0.0.0 --port 8840 --watcher=native

[group('docs')]
docs-build:
    cd mdbook && mdbook build -d ../docs
