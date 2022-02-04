import React, { Fragment, FC } from "react";

import headerImg from "../assets/logo.png";

const Header: FC = () => {
  return (
    <Fragment>
      <section className="about">
        <div className="container">
          <img src={headerImg} alt="" />
        </div>
      </section>
    </Fragment>
  );
};

export default Header;
