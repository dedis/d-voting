import { FC, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

const USERS_PER_PAGE = 5;

type UserIDTableProps = {
  userIDs: string[];
};

const UserIDTable: FC<UserIDTableProps> = ({ userIDs }) => {
  const { t } = useTranslation();

  const [userToDisplay, setUserToDisplay] = useState([]);
  const [pageIndex, setPageIndex] = useState(0);

  const partitionArray = (array: string[], size: number) =>
    array.map((_v, i) => (i % size === 0 ? array.slice(i, i + size) : null)).filter((v) => v);

  useEffect(() => {
    if (userIDs.length) {
      setUserToDisplay(partitionArray(userIDs, USERS_PER_PAGE)[pageIndex]);
    }
  }, [userIDs, pageIndex]);

  const handlePrevious = (): void => {
    if (pageIndex > 0) {
      setPageIndex(pageIndex - 1);
    }
  };

  const handleNext = (): void => {
    if (partitionArray(userIDs, USERS_PER_PAGE).length > pageIndex + 1) {
      setPageIndex(pageIndex + 1);
    }
  };

  return (
    <div>
      <div className="relative divide-y overflow-x-auto shadow-md sm:rounded-lg mt-2">
        <table className="w-full text-sm text-left text-gray-500">
          <thead className="text-xs text-gray-700 uppercase bg-gray-50">
            <tr>
              <th scope="col" className="px-6 py-3">
                User ID
              </th>
            </tr>
          </thead>
          <tbody>
            {userToDisplay !== undefined &&
              userToDisplay.map((user) => (
                <tr key={user} className="bg-white border-b">
                  <td className="px-1.5 sm:px-6 py-4 font-medium text-gray-900 whitespace-nowrap truncate">
                    {user}
                  </td>
                </tr>
              ))}
          </tbody>
        </table>

        <nav
          className="bg-white px-4 py-3 flex items-center justify-between border-t border-gray-200 sm:px-6"
          aria-label="Pagination">
          <div className="hidden sm:block text-sm text-gray-700">
            {t('showingNOverMOfXResults', {
              n: pageIndex + 1,
              m: partitionArray(userIDs, USERS_PER_PAGE).length,
              x: userIDs.length,
            })}
          </div>
          <div className="flex-1 flex justify-between sm:justify-end">
            <button
              disabled={pageIndex === 0}
              onClick={handlePrevious}
              className="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
              {t('previous')}
            </button>
            <button
              disabled={partitionArray(userIDs, USERS_PER_PAGE).length <= pageIndex + 1}
              onClick={handleNext}
              className="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
              {t('next')}
            </button>
          </div>
        </nav>
      </div>
    </div>
  );
};
export default UserIDTable;
