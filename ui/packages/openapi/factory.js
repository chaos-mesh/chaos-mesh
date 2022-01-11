import { cleanMarkers } from './utils.js'
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

        return
      }

      if (isArrayString(declaration.type, sourceFile)) {
        objs.push(_nodeToField(escapedName, declaration.type, declaration.jsDoc[0].comment ?? '', sourceFile))

        return
      }

      objs.push(typeReferenceToObjectLiteralExpression(escapedName, declaration.type, [], sourceFile, checker))

      return
    }

    objs.push(_nodeToField(escapedName, declaration.type, declaration.jsDoc[0].comment ?? '', sourceFile))
  })

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
    ],
    true
  )
}
