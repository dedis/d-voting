import React, { FC } from 'react';
import { Link } from 'react-router-dom';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

import useFetchCall from '../../../components/utils/useFetchCall';
import { ENDPOINT_EVOTING_GET_ALL } from '../../../components/utils/Endpoints';
import { GetAllElections, LightElectionInfo } from 'types/frontendRequestBody';

type SimpleTableProps = {
  statusToKeep: number;
  pathLink: string;
  textWhenData: string;
  textWhenNoData: string;
};

// Functional component that fetches all the elections, only keeps the elections
// whose status = statusToKeep and display them in a table with a single title
// column. It adds a link to '/pathLink/:id' when the title is clicked
// If table is empty, it display textWhenNoData instead
const SimpleTable: FC<SimpleTableProps> = ({
  statusToKeep,
  pathLink,
  textWhenData,
  textWhenNoData,
}) => {
  const { t } = useTranslation();
  const token = sessionStorage.getItem('token');
  const fetchRequest = {
    method: 'POST',
    body: JSON.stringify({ Token: token }),
  };
  const [fetchedData, loading, error] = useFetchCall(ENDPOINT_EVOTING_GET_ALL, fetchRequest);

  const ballotsToDisplay = (elections: GetAllElections) => {
    let dataToDisplay = [];
    elections.forEach((election) => {
      if (election.Status === statusToKeep) {
        dataToDisplay.push(election);
      }
    });
    return dataToDisplay;
  };

  const displayBallotTable = (data: LightElectionInfo[]) => {
    if (data.length > 0) {
      return (
        <div className="flex flex-col content-center items-center">
          <div className="vote-allowed mx-4 mt-8 mb-4">{textWhenData}</div>
          <div className="w-5/6 relative overflow-x-auto shadow-md sm:rounded-lg">
            <table className="w-full text-sm text-left text-gray-500 dark:text-gray-400">
              <thead className="text-xs text-gray-500 uppercase bg-gray-200">
                <tr>
                  <th scope="col" className="px-6 py-3">
                    {t('elecName')}
                  </th>
                </tr>
              </thead>
              <tbody>
                {data.map((election) => {
                  return (
                    <tr
                      className="block bg-white border-b  hover:bg-gray-50 "
                      key={election.ElectionID}>
                      <th
                        scope="row"
                        className="px-6 py-4 font-medium text-gray-500  whitespace-nowrap">
                        <Link
                          className="block text-gray-500"
                          to={{
                            pathname: `${pathLink}/${election.ElectionID}`,
                          }}>
                          {election.Title}
                        </Link>
                      </th>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      );
    } else {
      return <div>{textWhenNoData}</div>;
    }
  };

  const showBallots = (elections) => {
    return displayBallotTable(ballotsToDisplay(elections));
  };

  return (
    <div>
      {!loading ? (
        showBallots(fetchedData.AllElectionsInfo)
      ) : error === null ? (
        <p className="loading">{t('loading')}</p>
      ) : (
        <div className="error-retrieving">{t('errorRetrievingElection')}</div>
      )}
    </div>
  );
};

SimpleTable.propTypes = {
  statusToKeep: PropTypes.number.isRequired,
  pathLink: PropTypes.string.isRequired,
  textWhenData: PropTypes.string.isRequired,
  textWhenNoData: PropTypes.string.isRequired,
};

export default SimpleTable;
