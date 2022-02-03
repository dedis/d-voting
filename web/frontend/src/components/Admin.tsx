import { useState } from "react";
import Button from "@mui/material/Button";
import { DataGrid } from "@mui/x-data-grid";

import { GET_ADMIN_ROWS } from "./utils/ExpressEndoints";
import AddAdminUserModal from "./modal/AddAdminUserModal";
import RemoveAdminUserModal from "./modal/RemoveAdminUserModal";
import "../styles/Admin.css";

const Admin = () => {
  const [rows, setRows] = useState(undefined);
  const [newusrOpen, setNewusrOpen] = useState(false);

  const [sciperToDelete, setSciperToDelete] = useState(0);
  const [showDeleteModal, setShowDeleteModal] = useState(false);

  const openModal = () => setNewusrOpen(true);

  if (rows === undefined) {
    try {
      fetch(GET_ADMIN_ROWS)
        .then((resp) => {
          const json_data = resp.json();
          json_data.then((result) => {
            console.log(result);
            setRows(result);
          });
        })
        .catch((error) => {
          console.log(error);
        });
    } catch (error) {
      console.log(error);
    }
  }

  const columns = [
    {
      field: "sciper",
      headerName: "sciper",
      width: 150,
    },
    {
      field: "role",
      headerName: "role",
      width: 150,
    },
    {
      field: "action",
      headerName: "Action",
      width: 150,
      // eslint-disable-next-line react/display-name
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
        <AddAdminUserModal
          open={newusrOpen}
          setOpen={setNewusrOpen}
        ></AddAdminUserModal>
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
