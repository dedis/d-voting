import * as React from 'react';
import Button from '@mui/material/Button';
import {useContext, useState} from 'react';
import {DataGrid, GridRowsProp, GridColDef, GridApi, GridCellValue } from '@mui/x-data-grid';
import './Admin.css';
import {GET_ADMIN_ROWS} from '../utils/ExpressEndoints';
import {Translations} from '../language/Translations';
import {LanguageContext} from '../language/LanguageContext';
import BasicModal from '../modal/AddAdminUserModal'


const Admin = () => {
    const [context, ] = useContext(LanguageContext)
    const [rows, setRows] = useState(undefined);
    const [newusrOpen, setNewusrOpen] = useState(false);

    const openModal = () => setNewusrOpen(true);

    if(rows == undefined){
        try{
            fetch(GET_ADMIN_ROWS).then(resp => {
                const json_data = resp.json();
                json_data.then(result => {
                    console.log(result);
                    setRows(result);
                });
            }).catch(error => {
                console.log(error);
            });
        } catch (error){
            console.log(error);
        }
    }

    const columns = [
        {
            field: 'sciper',
            headerName: 'sciper',
            width: 150
        },
        {
            field: 'role',
            headerName: 'role',
            width: 150
        },
        {
            field: 'action',
            headerName: 'Action',
            width: 150,
            // eslint-disable-next-line react/display-name
            renderCell: function (params){
                function handledClick(){
                    console.log(params.id);
                }
                return <Button onClick={handledClick} variant="outlined" color="error">Delete</Button>
            },
        },
    ];

    return(
        <div className='admin-container'>
            <div className='admin-grid'>
                <Button onClick={openModal} variant="contained">Add a user</Button>
                <DataGrid rows={rows} columns={columns} />
                <BasicModal open={newusrOpen} setOpen={setNewusrOpen}></BasicModal>
            </div>
        </div>
    );
}

export default Admin;
