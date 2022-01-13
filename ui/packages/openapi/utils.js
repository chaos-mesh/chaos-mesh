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

const UI_FORM_ENUM = /ui:form:enum=(.+)\s/
const KUBEBUILDER_VALIDATION_ENUM = /\+kubebuilder:validation:Enum=(.+)\s/

/**
 * Get enum array from jsdoc comment.
 *
 * @export
 * @param {string} s
 * @return {string[]}
 */
export function getUIFormEnum(s) {
  let matched = s.match(UI_FORM_ENUM) || s.match(KUBEBUILDER_VALIDATION_ENUM)

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
