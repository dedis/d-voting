import React, { FC, useState } from 'react';
import { Link } from 'react-router-dom';
import Paper from '@material-ui/core/Paper';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableContainer from '@material-ui/core/TableContainer';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import TablePagination from '@material-ui/core/TablePagination';
import { withStyles } from '@material-ui/core/styles';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

import Action from './Action';
import Status from './Status';
//import ElectionFields from 'components/utils/ElectionFields';
import { LightElectionInfo } from 'types/electionInfo';
import { ID } from 'types/configuration';
import ElectionFields from 'components/utils/ElectionFields';

type ElectionTableProps = {
  elections: LightElectionInfo[];
};

// Returns a table where each line corresponds to an election with
// its name and status

const ElectionTable: FC<ElectionTableProps> = ({ elections }) => {
  const { t } = useTranslation();
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(5);

  const columns: Array<{
    id: string;
    label: string;
    minWidth: number;
    align: 'left';
  }> = [
    {
      id: 'title',
      label: t('title'),
      minWidth: 170,
      align: 'left',
    },
    {
      id: 'status',
      label: t('status'),
      minWidth: 170,
      align: 'left',
    },
    {
      id: 'action',
      label: t('action'),
      minWidth: 170,
      align: 'left',
    },
  ];

  const StyledTableRow = withStyles((theme) => ({
    root: {
      '&:nth-of-type(odd)': {
        backgroundColor: theme.palette.action.hover,
      },
    },
  }))(TableRow);

  const createData = (title: JSX.Element, status: JSX.Element, action: JSX.Element, key: ID) => {
    return { title, status, action, key };
  };

  const constructRows = () =>
    elections.map((election) => {
      let { title, id, status, setStatus } = ElectionFields(election);
      console.log('status:', status);
      let link = (
        <Link className="election-link" to={`/elections/${id}`}>
          {title}
        </Link>
      );
      let stat = <Status status={status} />;
      let action = <Action status={status} electionID={id} setStatus={setStatus} />;
      return createData(link, stat, action, id);
    });

  const rows = constructRows();
  const renderTH = () => {
    return (
      <TableRow className="row-head">
        {columns.map((col) => {
          return (
            <TableCell style={{ width: 800 }} key={col.id} align={col.align}>
              {col.label}
            </TableCell>
          );
        })}
      </TableRow>
    );
  };

  const handleChangePage = (event, newPage) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (event) => {
    setRowsPerPage(parseInt(event.target.value, 10));
    setPage(0);
  };

  return (
    <div>
      <Paper>
        <TableContainer>
          <Table>
            <TableHead className="table-header">{renderTH()}</TableHead>
            <TableBody>
              {rows.slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage).map((row) => {
                return (
                  <StyledTableRow key={row.key}>
                    {columns.map((column) => {
                      const value = row[column.id];
                      return (
                        <TableCell key={column.id} align={column.align}>
                          {value}
                        </TableCell>
                      );
                    })}
                  </StyledTableRow>
                );
              })}
            </TableBody>
          </Table>
        </TableContainer>
        <TablePagination
          rowsPerPageOptions={[5, 10, 25]}
          component="div"
          count={rows.length}
          rowsPerPage={rowsPerPage}
          page={page}
          onPageChange={handleChangePage}
          onChangeRowsPerPage={handleChangeRowsPerPage}
          labelDisplayedRows={({ from, to, count }) => {
            return '' + from + '-' + to + t('of') + count;
          }}
          labelRowsPerPage={t('rowsPerPage')}
        />
      </Paper>
    </div>
  );
};

ElectionTable.propTypes = {
  elections: PropTypes.array,
};

export default ElectionTable;
