/**
 * This file was auto-generated by @ui/openapi.
 * Do not make direct changes to the file.
 */

export const actions = ['error', 'random'],
  data = [
    {
      field: 'label',
      label: 'containerNames',
      value: [],
      helperText:
        'Optional. ContainerNames indicates list of the name of affected container. If not set, the first container will be injected',
    },
    {
      field: 'label',
      label: 'patterns',
      value: [],
      helperText:
        'Optional. Choose which domain names to take effect, support the placeholder ? and wildcard *, or the Specified domain name. Note:      1. The wildcard * must be at the end of the string. For example, chaos-*.org is invalid.      2. if the patterns is empty, will take effect on all the domain names. For example:   The value is [\\"google.com\\", \\"github.*\\", \\"chaos-mes?.org\\"],   will take effect on \\"google.com\\", \\"github.com\\" and \\"chaos-mesh.org\\"',
    },
  ]
