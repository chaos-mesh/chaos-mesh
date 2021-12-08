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
import Paper from '../Paper'

/**
 * PaperContainer usually be used to replace the default container.
 *
 * For example:
 *
 * <TableContainer component={PaperContainer}>
 * ...
 * </TableContainer>
 *
 * @param {React.ReactNode} { children }
 */
const PaperContainer: React.FC = ({ children }) => (
  <Paper sx={{ maxHeight: 768, p: 0, overflow: 'scroll' }}>{children}</Paper>
)

export default PaperContainer
