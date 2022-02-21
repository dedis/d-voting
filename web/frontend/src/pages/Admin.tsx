import React, { useState, useEffect } from 'react';
import Button from '@mui/material/Button';
import { DataGrid } from '@mui/x-data-grid';

import { ENDPOINT_USER_RIGHTS } from '../components/utils/Endpoints';
import AddAdminUserModal from '../components/modal/AddAdminUserModal';
import RemoveAdminUserModal from '../components/modal/RemoveAdminUserModal';
import './Admin.css';

const Admin = () => {
  const [rows, setRows] = useState([]);
  const [newusrOpen, setNewusrOpen] = useState(false);

  const [sciperToDelete, setSciperToDelete] = useState(0);
  const [showDeleteModal, setShowDeleteModal] = useState(false);

  const openModal = () => setNewusrOpen(true);

  useEffect(() => {
    if (newusrOpen || showDeleteModal) {
      return;
    }

    fetch(ENDPOINT_USER_RIGHTS)
      .then((resp) => {
        const jsonData = resp.json();
        jsonData.then((result) => {
          console.log(result);
          setRows(result);
        });
      })
      .catch((error) => {
        console.log(error);
      });
  }, [newusrOpen, showDeleteModal]);

  const columns = [
    {
      field: 'sciper',
      headerName: 'sciper',
      width: 150,
    },
    {
      field: 'role',
      headerName: 'role',
      width: 150,
    },
    {
      field: 'action',
      headerName: 'Action',
      width: 150,
      renderCell: function (params: any) {
        function handledClick() {
          setSciperToDelete(params.row.sciper);
          setShowDeleteModal(true);
        }
        return (
          <Button onClick={handledClick} variant="outlined" color="error">
            Delete
          </Button>
        );
      },
    },
  ];

  return (
    <div className="admin-container">
      <div className="admin-grid">
        <Button onClick={openModal} variant="contained">
          Add a user
        </Button>
        <DataGrid rows={rows} columns={columns} />
        <AddAdminUserModal open={newusrOpen} setOpen={setNewusrOpen}></AddAdminUserModal>
        <RemoveAdminUserModal
          setOpen={setShowDeleteModal}
          open={showDeleteModal}
          sciper={sciperToDelete}
        />
      </div>
    </div>
  );
};

export default Admin;
