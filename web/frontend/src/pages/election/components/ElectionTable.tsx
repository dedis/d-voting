import React, { FC, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';
import { LightElectionInfo } from 'types/election';
import Status from './Status';
import QuickAction from './QuickAction';
import ElectionTableFilter from './ElectionTableFilter';
import ResetFilterButton from './ResetFilterButton';

type ElectionTableProps = {
  elections: LightElectionInfo[];
};

// Returns a table where each line corresponds to an election with
// its name and status
const ELECTION_PER_PAGE = 10;

const ElectionTable: FC<ElectionTableProps> = ({ elections }) => {
  const { t } = useTranslation();
  const [pageIndex, setPageIndex] = useState(0);
  const [electionsToDisplay, setElectionsToDisplay] = useState<LightElectionInfo[]>([]);
  const [statusToKeep, setStatusToKeep] = useState(null);

  const displayAllElections = () => {
    if (elections) {
      const electToDisplay = [];
      elections.forEach((election) => {
        electToDisplay.push(election);
      });

      setElectionsToDisplay(electToDisplay);
    }
  };

  useEffect(() => {
    displayAllElections();
  }, [elections]);

  const partitionArray = (array: LightElectionInfo[], size: number) =>
    array.map((v, i) => (i % size === 0 ? array.slice(i, i + size) : null)).filter((v) => v);

  useEffect(() => {
    if (statusToKeep) {
      const newElectionsToDisplay = elections.filter(
        (election) => election.Status === statusToKeep
      );
      setElectionsToDisplay(newElectionsToDisplay);
    } else {
      displayAllElections();
    }
  }, [statusToKeep]);

  useEffect(() => {
    if (electionsToDisplay.length) {
      setElectionsToDisplay(partitionArray(electionsToDisplay, ELECTION_PER_PAGE)[pageIndex]);
    }
  }, [pageIndex]);

  const handlePrevious = (): void => {
    if (pageIndex > 0) {
      setPageIndex(pageIndex - 1);
    }
  };

  const handleNext = (): void => {
    if (partitionArray(electionsToDisplay, ELECTION_PER_PAGE).length > pageIndex + 1) {
      setPageIndex(pageIndex + 1);
    }
  };

  useEffect(() => {
    if (electionsToDisplay.length) {
      setElectionsToDisplay(partitionArray(electionsToDisplay, ELECTION_PER_PAGE)[pageIndex]);
    }
  }, [pageIndex]);

  return (
    <div>
      <div className="relative divide-y overflow-x-auto shadow-md sm:rounded-lg">
        <table className="w-full text-sm text-left text-gray-500">
          <thead className="text-xs text-gray-700 uppercase bg-gray-50">
            <tr>
              <th scope="col" className="px-6 py-3">
                {t('elecName')}
              </th>
              <th scope="col" className="px-6 py-3">
                {t('status')}
              </th>
              <th scope="col" className="px-6 py-3">
                <span className="sr-only">Edit</span>
              </th>
            </tr>
          </thead>
          <tbody>
            {electionsToDisplay.map((election) => (
              <tr key={election.ElectionID} className="bg-white border-b hover:bg-gray-50 ">
                <td scope="row" className="px-6 py-4 font-medium text-gray-900 whitespace-nowrap">
                  <Link
                    className="election-link text-gray-700 hover:text-indigo-500"
                    to={`/elections/${election.ElectionID}`}>
                    {election.Title}
                  </Link>
                </td>
                <td className="px-6 py-4">
                  <Status status={election.Status} />
                </td>
                <td className="px-6 py-4 text-right">
                  <QuickAction status={election.Status} electionID={election.ElectionID} />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        <nav
          className="bg-white px-4 py-3 flex items-center justify-between border-t border-gray-200 sm:px-6"
          aria-label="Pagination">
          <div className="hidden sm:block">
            <p className="text-sm text-gray-700">
              {t('showing')} <span className="font-medium">{pageIndex + 1}</span> /{' '}
              <span className="font-medium">
                {partitionArray(electionsToDisplay, ELECTION_PER_PAGE).length}
              </span>{' '}
              {t('of')} <span className="font-medium">{electionsToDisplay.length}</span>{' '}
              {t('results')}
            </p>
          </div>
          <div className="flex-1 flex justify-between sm:justify-end">
            <button
              disabled={pageIndex === 0}
              onClick={handlePrevious}
              className="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
              {t('previous')}
            </button>
            <button
              disabled={
                partitionArray(electionsToDisplay, ELECTION_PER_PAGE).length <= pageIndex + 1
              }
              onClick={handleNext}
              className="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
              {t('next')}
            </button>
          </div>
        </nav>
      </div>
      <ElectionTableFilter setStatusToKeep={setStatusToKeep} />
      <ResetFilterButton setStatusToKeep={setStatusToKeep} />
    </div>
  );
};

ElectionTable.propTypes = {
  elections: PropTypes.array,
};

export default ElectionTable;
