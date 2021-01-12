SRC_API=src/api/

cd $SRC_API

`npm bin`/ts-interface-builder archives.type.ts --inline-imports
`npm bin`/ts-interface-builder common.type.ts --inline-imports

cd -
