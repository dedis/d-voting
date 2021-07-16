import {React, useContext} from 'react';
import {Translations} from '../language/Translations';
import {LanguageContext} from '../language/LanguageContext';
import useFetchCall from '../utils/useFetchCall';
import {GET_ALL_ELECTIONS_ENDPOINT} from '../utils/Endpoints'
import {Link} from 'react-router-dom';
import Paper from '@material-ui/core/Paper';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableContainer from '@material-ui/core/TableContainer';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import { withStyles} from '@material-ui/core/styles';
import PropTypes from 'prop-types';

/*functional component that fetches all the elections, only keeps the elections
whose status = statusToKeep and display them in a table with a single title
column. It adds a link to '/pathLink/:id' when the title is clicked 
If table is empty, it display textWhenNoData instead*/
const SimpleTable = ({statusToKeep,pathLink, textWhenData, textWhenNoData}) => {
    const [context, ] = useContext(LanguageContext);
    const token = sessionStorage.getItem('token');
    const fetchRequest = {
        method: 'POST',
        body: JSON.stringify({'Token': token})
    }
    const [data, loading, error] = useFetchCall(GET_ALL_ELECTIONS_ENDPOINT, fetchRequest);
    const StyledTableRow = withStyles((theme) => ({
        root: {
          '&:nth-of-type(odd)': {
            backgroundColor: theme.palette.action.hover,
          },
        },
      }))(TableRow);
    
    const ballotsToDisplay = (elections) => {
       let dataToDisplay = [];
       elections.map((elec) => {
           if(elec.Status === statusToKeep){
               dataToDisplay.push([elec.Title, elec.ElectionID]);
           }
       })
       return dataToDisplay;
   }

    const displayBallotTable = (data) => {
        if(data.length > 0){
            return (
                <div>
                 <div className='vote-allowed'>{textWhenData}</div>   
                <Paper>
                    <TableContainer>
                        <Table stickyHeader aria-label = "sticky table">
                            <TableHead className = 'table-header'>
                                <TableRow className='row-head'>
                                    <TableCell key = {'Title'}>
                                        {Translations[context].elecName}
                                    </TableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                {data.map((row) => {
                                    return(
                                        <StyledTableRow key={row}>
                                            <TableCell key = {row[1]}>
                                                <Link className='election-link' to={{pathname:`/${pathLink}/${row[1]}`,
                                                data: row[1]}}>{row[0]}</Link>
                                            </TableCell>
                                        </StyledTableRow>
                                    )
                                })}
                            </TableBody>
                        </Table>
                    </TableContainer>
                </Paper>
            </div>);
        } else {
            return <div>{textWhenNoData}</div>;
        }
    }
    const showBallots = (elections) => {
        return (
            displayBallotTable(ballotsToDisplay(elections))
        )}

    return (
        <div className = 'cast-ballot'>
            {!loading? showBallots(data.AllElectionsInfo):
                (error === null?<p className='loading'>{Translations[context].loading}</p>:<div className='error-retrieving'>{Translations[context].errorRetrievingElection}</div>)          
            }    
        </div>
    )
}

SimpleTable.propTypes = {
    statusToKeep : PropTypes.number.isRequired,
    pathLink: PropTypes.string.isRequired,
    textWhenData: PropTypes.string.isRequired,
    textWhenNoData: PropTypes.string.isRequired,

}

export default SimpleTable;