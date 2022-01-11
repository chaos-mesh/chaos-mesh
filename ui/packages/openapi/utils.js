const UI_FORM_ENUM = /ui:form:enum=(.+)\s/

/**
 * Get enum array from jsdoc comment.
 *
 * @export
 * @param {string} s
 * @return {string[]}
 */
export function getUIFormEnum(s) {
  const matched = s.match(UI_FORM_ENUM)

  return matched ? matched[1].split(';') : []
}

const UI_FORM_ACTION = /ui:form:action=(.+)\s/

/**
 * Get action name from jsdoc comment. If not found, return an empty string.
 *
 * @export
 * @param {string} s
 * @return {string}
 */
export function getUIFormAction(s) {
  const matched = s.match(UI_FORM_ACTION)

  return matched ? matched[1] : ''
}

const UI_FORM_IGNORE = /ui:form:ignore\s/

/**
 * Determine if jsdoc comment contains the ignored keyword.
 *
 * @export
 * @param {string} s
 * @return {boolean}
 */
export function isUIFormIgnore(s) {
  return UI_FORM_IGNORE.test(s)
}

/**
 * Remove markers(ui:form..., +kubebuilder+..., +optional..., etc.) from jsdoc comment.
 *
 * @export
 * @param {string} s
 */
export function cleanMarkers(s) {
  s = s.replace(UI_FORM_ACTION, '')

  const reOptional = /\+optional/
  if (reOptional.test(s)) {
    s = 'Optional. ' + s.replace(reOptional, '')
  }

  s = s.replace(/\+kubebuilder\S+\s/, '').trim()

  return s
}
