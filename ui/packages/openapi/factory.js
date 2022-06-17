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
import ts from 'typescript'

import { cleanMarkers, getUIFormWhen, isUIFormIgnore } from './utils.js'

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
    case 'number':
      return 'number'
    case 'boolean':
      return 'select'
    case 'Array<string>':
      return 'label'
    case 'Array<number>':
      return 'numbers'
    case '{ [key: string]: string }':
      return 'text-text'
    case '{ [key: string]: Array<string> }':
      return 'text-label'
    default:
      throw new Error(`Unsupported type: ${type}`)
  }
}

/**
 * Convert typescript internal type to form field's intial value.
 *
 * @param {string} type
 * @return {ts.StringLiteral|ts.ArrayLiteralExpression|ts.ObjectLiteralExpression|ts.NumericLiteral|null}
 */
export function typeTextToInitialValue(type) {
  switch (type) {
    case 'string':
      return factory.createStringLiteral('')
    case 'number':
      return factory.createNumericLiteral(0)
    case 'boolean':
      return factory.createFalse()
    case 'Array<string>':
    case 'Array<number>':
      return factory.createArrayLiteralExpression()
    case '{ [key: string]: string }':
    case '{ [key: string]: Array<string> }':
      return factory.createObjectLiteralExpression()
    default:
      throw new Error(`Unsupported type: ${type}`)
  }
}

/**
 * Check if the type is a array literal.
 *
 * @param {any} type
 * @param {ts.sourceFile} sourceFile
 * @return {boolean}
 */
function isArrayLiteral(type, sourceFile) {
  /** @type {string} */
  const typeText = type.getText(sourceFile)

  return typeText.startsWith('Array<')
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
  if (type.kind === ts.SyntaxKind.TypeReference && !isArrayLiteral(type, sourceFile)) {
    return typeReferenceToObjectLiteralExpression(identifier, type, comment, objs, sourceFile, checker)
  }

  return _nodeToField(identifier, type, comment, sourceFile)
}

/**
 * TypeReference to ObjectLiteralExpression.
 *
 * @param {string} identifier
 * @param {ts.TypeReference} typeRef
 * @param {string} comment
 * @param {ts.Expression[]} objs - usually an empty array
 * @param {ts.SourceFile} sourceFile
 * @param {ts.TypeChecker} checker
 * @param {object} options
 * @param {boolean} [options.multiple] - if true, the field is an array
 */
function typeReferenceToObjectLiteralExpression(
  identifier,
  typeRef,
  comment,
  objs,
  sourceFile,
  checker,
  options = { multiple: false }
) {
  const type = checker.getTypeAtLocation(typeRef)
  const when = getUIFormWhen(comment)
  const members = type.symbol.members

  members.forEach((val) => {
    const { escapedName, valueDeclaration: declaration } = val

    const comment = declaration.jsDoc[0].comment ?? ''
    if (isUIFormIgnore(comment)) {
      return
    }

    if (ts.isTypeReferenceNode(declaration.type)) {
      if (isHTTTPChaosPatchHeadersOrQueries(declaration)) {
        objs.push(
          factory.createObjectLiteralExpression(
            _genBaseFieldElements(escapedName, '{ [key: string]: Array<string> }', comment),
            true
          )
        )

        return
      }

      // handle non-primritive array
      if (
        declaration.type.typeName.escapedText === 'Array' &&
        declaration.type.typeArguments[0].kind === ts.SyntaxKind.TypeReference
      ) {
        objs.push(
          typeReferenceToObjectLiteralExpression(
            escapedName,
            declaration.type.typeArguments[0],
            comment,
            [],
            sourceFile,
            checker,
            { multiple: true }
          )
        )
      } else if (isArrayLiteral(declaration.type, sourceFile)) {
        // handle literal array
        objs.push(_nodeToField(escapedName, declaration.type, comment, sourceFile))
      } else {
        objs.push(
          typeReferenceToObjectLiteralExpression(escapedName, declaration.type, comment, [], sourceFile, checker)
        )
      }

      return
    }

    objs.push(_nodeToField(escapedName, declaration.type, comment, sourceFile))
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
      ...(when
        ? [factory.createPropertyAssignment(factory.createIdentifier('when'), factory.createStringLiteral(when))]
        : []),
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

  const properties = [
    ..._genBaseFieldElements(identifier, typeText, comment),
    ...(when
      ? [factory.createPropertyAssignment(factory.createIdentifier('when'), factory.createStringLiteral(when))]
      : []),
  ]
  // {
  //   ..._genBaseFieldElements(),
  //   when?: '',
  // }
  return factory.createObjectLiteralExpression(properties, true)
}

/**
 *
 *
 * @param {string} identifier
 * @param {string} typeText
 * @param {string} comment
 * @return {ts.PropertyAssignment[]}
 */
function _genBaseFieldElements(identifier, typeText, comment) {
  // {
  //   field: '',
  //   label: '',
  //   value: '',
  //   items: [],
  //   helperText: '',
  // }
  return [
    factory.createPropertyAssignment(
      factory.createIdentifier('field'),
      factory.createStringLiteral(typeTextToFieldType(typeText))
    ),
    factory.createPropertyAssignment(factory.createIdentifier('label'), factory.createStringLiteral(identifier)),
    factory.createPropertyAssignment(factory.createIdentifier('value'), typeTextToInitialValue(typeText)),
    ...(typeText === 'boolean'
      ? [
          factory.createPropertyAssignment(
            factory.createIdentifier('items'),
            factory.createArrayLiteralExpression([factory.createTrue(), factory.createFalse()])
          ),
        ]
      : []),
    factory.createPropertyAssignment(
      factory.createIdentifier('helperText'),
      factory.createStringLiteral(cleanMarkers(comment))
    ),
  ]
}

/**
 * Find special identifiers `headers` and `queries` in HTTPChaos Patch.
 * The type of them is `Array<Array<string>>`.
 *
 * @export
 * @param {ts.Node} node
 */
export function isHTTTPChaosPatchHeadersOrQueries(node) {
  const identifier = node.name.escapedText

  return (
    node.parent.name.escapedText === 'V1alpha1PodHttpChaosPatchActions' &&
    (identifier === 'headers' || identifier === 'queries')
  )
}
