/*
 * Copyright 2021 Chaos Mesh Authors.
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

/**
 * Checks if an experiment kind is enabled
 * @param kind - The experiment kind to check
 * @param enabledExperiments - Optional list of enabled experiment kinds from API config
 */
export function isExperimentEnabled(kind: string, enabledExperiments?: string[]): boolean {
  if (!enabledExperiments || enabledExperiments.length === 0) {
    return true
  }

  return enabledExperiments.includes(kind)
}

/**
 * Checks if any physical machine experiments are enabled
 * Physical machine experiments include:
 * - PhysicalMachineChaos (for workflows)
 * - DiskChaos, NetworkChaos, TimeChaos, JVMChaos, ProcessChaos (when in physic env)
 * @param enabledExperiments - Optional list of enabled experiment kinds from API config
 */
export function hasPhysicalMachineExperimentsEnabled(enabledExperiments?: string[]): boolean {
  // If no experiments are specified, physical experiments are enabled (backward compatibility)
  if (!enabledExperiments || enabledExperiments.length === 0) {
    return true
  }

  // Physical machine experiment kinds that can appear in the list
  const physicalKinds = ['PhysicalMachineChaos', 'DiskChaos', 'ProcessChaos']

  return physicalKinds.some((kind) => enabledExperiments.includes(kind))
}
