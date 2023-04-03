/**
 * This file was auto-generated by @ui/openapi.
 * Do not make direct changes to the file.
 */

export const actions = [],
  data = [
    {
      field: 'select',
      label: 'abort',
      value: false,
      items: [true, false],
      helperText: 'Optional. Abort is a rule to abort a http session.',
    },
    {
      field: 'number',
      label: 'code',
      value: 0,
      helperText: 'Optional. Code is a rule to select target by http status code in response.',
    },
    {
      field: 'text',
      label: 'delay',
      value: '',
      helperText:
        'Optional. Delay represents the delay of the target request/response. A duration string is a possibly unsigned sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "2h45m". Valid time units are "ns", "us" (or "\u00B5s"), "ms", "s", "m", "h".',
    },
    {
      field: 'text',
      label: 'method',
      value: '',
      helperText: 'Optional. Method is a rule to select target by http method in request.',
    },
    {
      field: 'ref',
      label: 'patch',
      children: [
        {
          field: 'ref',
          label: 'body',
          children: [
            {
              field: 'text',
              label: 'type',
              value: '',
              helperText:
                'Type represents the patch type, only support `JSON` as [merge patch json](https://tools.ietf.org/html/rfc7396) currently.',
            },
            {
              field: 'text',
              label: 'value',
              value: '',
              helperText: 'Value is the patch contents.',
            },
          ],
        },
        {
          field: 'text-label',
          label: 'headers',
          value: {},
          helperText:
            'Optional. Headers is a rule to append http headers of target. For example: `[["Set-Cookie", "<one cookie>"], ["Set-Cookie", "<another cookie>"]]`.',
        },
        {
          field: 'text-label',
          label: 'queries',
          value: {},
          helperText:
            'Optional. Queries is a rule to append uri queries of target(Request only). For example: `[["foo", "bar"], ["foo", "unknown"]]`.',
        },
      ],
    },
    {
      field: 'text',
      label: 'path',
      value: '',
      helperText: 'Optional. Path is a rule to select target by uri path in http request.',
    },
    {
      field: 'number',
      label: 'port',
      value: 0,
      helperText: 'Port represents the target port to be proxy of.',
    },
    {
      field: 'text',
      label: 'remoteCluster',
      value: '',
      helperText: 'Optional. RemoteCluster represents the remote cluster where the chaos will be deployed',
    },
    {
      field: 'ref',
      label: 'replace',
      children: [
        {
          field: 'numbers',
          label: 'body',
          value: [],
          helperText: 'Optional. Body is a rule to replace http message body in target.',
        },
        {
          field: 'number',
          label: 'code',
          value: 0,
          helperText: 'Optional. Code is a rule to replace http status code in response.',
        },
        {
          field: 'text-text',
          label: 'headers',
          value: {},
          helperText:
            'Optional. Headers is a rule to replace http headers of target. The key-value pairs represent header name and header value pairs.',
        },
        {
          field: 'text',
          label: 'method',
          value: '',
          helperText: 'Optional. Method is a rule to replace http method in request.',
        },
        {
          field: 'text',
          label: 'path',
          value: '',
          helperText: 'Optional. Path is rule to to replace uri path in http request.',
        },
        {
          field: 'text-text',
          label: 'queries',
          value: {},
          helperText:
            'Optional. Queries is a rule to replace uri queries in http request. For example, with value `{ "foo": "unknown" }`, the `/?foo=bar` will be altered to `/?foo=unknown`,',
        },
      ],
    },
    {
      field: 'text-text',
      label: 'request_headers',
      value: {},
      helperText:
        'Optional. RequestHeaders is a rule to select target by http headers in request. The key-value pairs represent header name and header value pairs.',
    },
    {
      field: 'text-text',
      label: 'response_headers',
      value: {},
      helperText:
        'Optional. ResponseHeaders is a rule to select target by http headers in response. The key-value pairs represent header name and header value pairs.',
    },
    {
      field: 'text',
      label: 'target',
      value: '',
      helperText: 'Target is the object to be selected and injected.',
    },
    {
      field: 'ref',
      label: 'tls',
      children: [
        {
          field: 'text',
          label: 'caName',
          value: '',
          helperText: 'Optional. CAName represents the data name of ca file in secret, `ca.crt` for example',
        },
        {
          field: 'text',
          label: 'certName',
          value: '',
          helperText: 'CertName represents the data name of cert file in secret, `tls.crt` for example',
        },
        {
          field: 'text',
          label: 'keyName',
          value: '',
          helperText: 'KeyName represents the data name of key file in secret, `tls.key` for example',
        },
        {
          field: 'text',
          label: 'secretName',
          value: '',
          helperText: 'SecretName represents the name of required secret resource',
        },
        {
          field: 'text',
          label: 'secretNamespace',
          value: '',
          helperText: 'SecretNamespace represents the namespace of required secret resource',
        },
      ],
    },
  ]
