import React, { Fragment, FC } from "react";

import aboutImg from "../assets/dvoting_dela.png";

const Header: FC = () => {
  return (
    <Fragment>
      <section className="about">
        <div className="container">
          <img src={aboutImg} alt="" />
        </div>
      </section>
    </Fragment>
  );
};

export default About;
