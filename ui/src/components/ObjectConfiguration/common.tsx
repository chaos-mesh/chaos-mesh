import { ListItemProps, ListItem as MUIListItem, TableCell as MUITableCell, TableCellProps } from '@material-ui/core'

export const TableCell = (props: TableCellProps) => <MUITableCell sx={{ borderBottom: 'none' }} {...props} />
export const ListItem = (props: ListItemProps) => <MUIListItem sx={{ pl: 0 }} {...props} />
