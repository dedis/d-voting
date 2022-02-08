import * as React from "react";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Typography from "@mui/material/Typography";
import Modal from "@mui/material/Modal";
import { DELETE_API_ROLE } from "../utils/ExpressEndoints";
import Stack from "@mui/material/Stack";

const style = {
  position: "absolute",
  top: "50%",
  left: "50%",
  transform: "translate(-50%, -50%)",
  width: 400,
  bgcolor: "background.paper",
  border: "2px solid #000",
  boxShadow: 24,
  p: 4,
};

export default function RemoveAdminUserModal({ open, setOpen, sciper }) {
  const handleClose = () => setOpen(false);

  const handleDelete = () => {
    const requestOptions = {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ sciper: sciper }),
    };
    fetch(DELETE_API_ROLE, requestOptions).then((data) => {
      if (data.status === 200) {
        alert("User removed successfully");
        setOpen(false);
      } else {
        alert("Error while adding the user");
      }
    });
  };

  return (
    <div>
      <Modal
        open={open}
        onClose={handleClose}
        aria-labelledby="modal-title"
        aria-describedby="modal-description"
      >
        <Box sx={style}>
          <Typography variant="h6" component="h2">
            Please confirm deletion for sciper {sciper}
          </Typography>

          <Stack direction="row" spacing={2}>
            <Button onClick={handleDelete} variant="outlined">
              Delete
            </Button>
            <Button onClick={handleClose} variant="contained">
              Cancel
            </Button>
          </Stack>
        </Box>
      </Modal>
    </div>
  );
}
