SRC_API=src/api/

cd $SRC_API

`npm bin`/ts-interface-builder archives.type.ts --inline-imports

cd -
