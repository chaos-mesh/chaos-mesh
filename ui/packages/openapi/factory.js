/*
 * Copyright 2022 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

import { cleanMarkers, getUIFormWhen } from './utils.js'

import ts from 'typescript'

const { factory } = ts

/**
 * Convert typescript internal type to form field type.
 *
 * @param {string} type
 * @return {string}
 */
export function typeTextToFieldType(type) {
  switch (type) {
    case 'string':
      return 'text'
    case 'Array<string>':
    case 'string[]':
      return 'label'
    default:
      return 'text'
  }
}

/**
 * Convert typescript internal type to form field's intial value.
 *
 * @param {string} type
 * @return {ts.StringLiteral|ts.ArrayLiteralExpression|ts.NumericLiteral|null}
 */
export function typeTextToInitialValue(type) {
  switch (type) {
    case 'string':
      return ts.factory.createStringLiteral('')
    case 'Array<string>':
    case 'string[]':
      return ts.factory.createArrayLiteralExpression([], false)
    case 'number':
      return ts.factory.createNumericLiteral(0)
    default:
      return ts.factory.createStringLiteral('')
  }
}

/**
 * Check if the type is a array string.
 *
 * @param {any} type
 * @param {ts.sourceFile} sourceFile
 * @return {boolean}
 */
function isArrayString(type, sourceFile) {
  const typeText = type.getText(sourceFile)

  return typeText === 'Array<string>' || typeText === 'string[]'
}

/**
 * Generate form field.
 *
 * @export
 * @param {string} identifier
 * @param {any} type
 * @param {string} comment
 * @param {ts.Expression[]} objs - usually an empty array
 * @param {ts.SourceFile} sourceFile
 * @param {ts.TypeChecker} checker
 * @return {ts.ObjectLiteralExpression}
 */
export function nodeToField(identifier, type, comment, objs, sourceFile, checker) {
  // handle TypeReference
  if (type.kind === ts.SyntaxKind.TypeReference && !isArrayString(type, sourceFile)) {
    return typeReferenceToObjectLiteralExpression(identifier, type, objs, sourceFile, checker)
  }

  return _nodeToField(identifier, type, comment, sourceFile)
}

/**
 * TypeReference to ObjectLiteralExpression.
 *
 * @param {string} identifier
 * @param {ts.TypeReference} typeRef
 * @param {ts.Expression[]} objs - usually an empty array
 * @param {ts.SourceFile} sourceFile
 * @param {ts.TypeChecker} checker
 * @param {Object} options
 * @param {boolean} [options.multiple] - if true, the field is an array
 */
function typeReferenceToObjectLiteralExpression(
  identifier,
  typeRef,
  objs,
  sourceFile,
  checker,
  options = { multiple: false }
) {
  const type = checker.getTypeAtLocation(typeRef)
  const members = type.symbol.members

  members.forEach((val) => {
    const { escapedName, valueDeclaration: declaration } = val

    if (declaration.type.kind === ts.SyntaxKind.TypeReference) {
      // handle non-primritive array
      if (
        declaration.type.typeName.escapedText === 'Array' &&
        declaration.type.typeArguments[0].kind === ts.SyntaxKind.TypeReference
      ) {
        objs.push(
          typeReferenceToObjectLiteralExpression(
            escapedName,
            declaration.type.typeArguments[0],
            [],
            sourceFile,
            checker,
            { multiple: true }
          )
        )
      } else if (isArrayString(declaration.type, sourceFile)) {
        // handle string array
        objs.push(_nodeToField(escapedName, declaration.type, declaration.jsDoc[0].comment ?? '', sourceFile))
      } else {
        objs.push(typeReferenceToObjectLiteralExpression(escapedName, declaration.type, [], sourceFile, checker))
      }

      return
    }

    objs.push(_nodeToField(escapedName, declaration.type, declaration.jsDoc[0].comment ?? '', sourceFile))
  })

  // create ref field
  //
  // {
  //   field: 'ref',
  //   label: '',
  //   children: []
  // }
  return factory.createObjectLiteralExpression(
    [
      factory.createPropertyAssignment(factory.createIdentifier('field'), factory.createStringLiteral('ref')),
      factory.createPropertyAssignment(factory.createIdentifier('label'), factory.createStringLiteral(identifier)),
      ...(options.multiple
        ? [factory.createPropertyAssignment(factory.createIdentifier('multiple'), factory.createTrue())]
        : []),
      factory.createPropertyAssignment(
        factory.createIdentifier('children'),
        factory.createArrayLiteralExpression(objs, true)
      ),
    ],
    true
  )
}

/**
 * Generate atomic form field.
 *
 * @param {string} identifier
 * @param {any} type
 * @param {string} comment
 * @param {ts.sourceFile} sourceFile
 * @return {ts.ObjectLiteralExpression}
 */
function _nodeToField(identifier, type, comment, sourceFile) {
  const typeText = type.getText(sourceFile)
  const when = getUIFormWhen(comment)

  // {
  //   field: '',
  //   label: '',
  //   value: '',
  //   helperText: '',
  // }
  return factory.createObjectLiteralExpression(
    [
      factory.createPropertyAssignment(
        factory.createIdentifier('field'),
        factory.createStringLiteral(typeTextToFieldType(typeText))
      ),
      factory.createPropertyAssignment(factory.createIdentifier('label'), factory.createStringLiteral(identifier)),
      factory.createPropertyAssignment(factory.createIdentifier('value'), typeTextToInitialValue(typeText)),
      factory.createPropertyAssignment(
        factory.createIdentifier('helperText'),
        factory.createStringLiteral(cleanMarkers(comment))
      ),
      ...(when
        ? [factory.createPropertyAssignment(factory.createIdentifier('when'), factory.createStringLiteral(when))]
        : []),
    ],
    true
  )
}
