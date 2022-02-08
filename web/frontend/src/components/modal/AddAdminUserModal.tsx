import * as React from "react";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Typography from "@mui/material/Typography";
import Modal from "@mui/material/Modal";
import Input from "@mui/material/Input";
import { useState } from "react";
import { ADD_API_ROLE } from "../utils/ExpressEndoints";
import InputLabel from "@mui/material/InputLabel";
import MenuItem from "@mui/material/MenuItem";
import FormControl from "@mui/material/FormControl";
import Select from "@mui/material/Select";

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

export default function AddAdminUserModal({ open, setOpen }) {
  const handleClose = () => setOpen(false);
  const ariaLabel = { "aria-label": "description" };

  const [sciperValue, setSciperValue] = useState("");

  const [roleValue, setRoleValue] = useState("");

  const handleChange = (event: any) => {
    setRoleValue(event.target.value);
  };

  const handleUserInput = (e: any) => {
    setSciperValue(e.target.value);
  };

  const handleClick = () => {
    const requestOptions = {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ sciper: sciperValue, role: roleValue }),
    };
    fetch(ADD_API_ROLE, requestOptions).then((data) => {
      if (data.status === 200) {
        alert("User added successfully");
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
            Please give the sciper of the user
          </Typography>
          <Input
            value={sciperValue}
            onChange={handleUserInput}
            placeholder="Sciper"
            inputProps={ariaLabel}
          />
          <br />
          <br />
          <Box sx={{ minWidth: 40 }}>
            <FormControl fullWidth>
              <InputLabel id="select-label-role">Role</InputLabel>
              <Select
                labelId="select-label-role"
                id="select-role"
                value={roleValue}
                label="Role"
                onChange={handleChange}
              >
                <MenuItem value={"operator"}>Operator</MenuItem>
                <MenuItem value={"admin"}>Admin</MenuItem>
              </Select>
            </FormControl>
          </Box>
          <br />
          <Button onClick={handleClick} variant="contained">
            Add User
          </Button>
        </Box>
      </Modal>
    </div>
  );
}
