import { useContext, useEffect, useState } from 'react';

import { ENDPOINT_USER_RIGHTS } from 'components/utils/Endpoints';

import AddAdminUserModal from 'components/modal/AddAdminUserModal';
import { useTranslation } from 'react-i18next';
import RemoveAdminUserModal from 'components/modal/RemoveAdminUserModal';
import Loading from './Loading';
import { FlashContext, FlashLevel } from 'index';

const SCIPERS_PER_PAGE = 10;

const Admin = () => {
  const { t } = useTranslation();
  const fctx = useContext(FlashContext);

  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [newUserOpen, setNewUserOpen] = useState(false);
  const [scipersToDisplay, setScipersToDisplay] = useState([]);
  const [sciperToDelete, setSciperToDelete] = useState(0);
  const [pageIndex, setPageIndex] = useState(0);

  const openModal = () => setNewUserOpen(true);

  useEffect(() => {
    setLoading(true);
    fetch(ENDPOINT_USER_RIGHTS)
      .then((resp) => {
        setLoading(false);
        if (resp.status === 200) {
          const jsonData = resp.json();
          jsonData.then((result) => {
            setUsers(result);
          });
        } else {
          setUsers([]);
          fctx.addMessage(t('errorFetchingUsers'), FlashLevel.Error);
        }
      })
      .catch((error) => {
        setLoading(false);
        fctx.addMessage(`${t('errorFetchingUsers')}: ${error.message}`, FlashLevel.Error);
      });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const partitionArray = (array: any[], size: number) =>
    array.map((v, i) => (i % size === 0 ? array.slice(i, i + size) : null)).filter((v) => v);

  useEffect(() => {
    if (users.length) {
      setScipersToDisplay(partitionArray(users, SCIPERS_PER_PAGE)[pageIndex]);
    }
  }, [users, pageIndex]);

  const handleDelete = (sciper: number): void => {
    setSciperToDelete(sciper);
    setShowDeleteModal(true);
  };

  const handlePrevious = (): void => {
    if (pageIndex > 0) {
      setPageIndex(pageIndex - 1);
    }
  };
  const handleNext = (): void => {
    if (partitionArray(users, SCIPERS_PER_PAGE).length > pageIndex + 1) {
      setPageIndex(pageIndex + 1);
    }
  };

  const handleAddRoleUser = (user: object): void => {
    setUsers([...users, user]);
  };
  const handleRemoveRoleUser = (): void => {
    setUsers(users.filter((user) => user.sciper !== sciperToDelete));
  };

  return !loading ? (
    <div className="w-[60rem] font-sans px-4 py-8">
      <AddAdminUserModal
        open={newUserOpen}
        setOpen={setNewUserOpen}
        handleAddRoleUser={handleAddRoleUser}
      />
      <RemoveAdminUserModal
        setOpen={setShowDeleteModal}
        open={showDeleteModal}
        sciper={sciperToDelete}
        handleRemoveRoleUser={handleRemoveRoleUser}
      />
      <div className="flex items-center justify-between mb-4">
        <div className="flex-1 min-w-0">
          <h2 className="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
            {t('admin')}
          </h2>
          <div className="mt-1 flex flex-col sm:flex-row sm:flex-wrap sm:mt-0 sm:space-x-6">
            <div className="mt-2 flex items-center text-sm text-gray-500">{t('adminDetails')}</div>
          </div>
        </div>
        <div className="mt-5 flex lg:mt-0 lg:ml-4">
          <span className="sm:ml-3">
            <button
              type="button"
              onClick={openModal}
              className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700">
              {t('addUser')}
            </button>
          </span>
        </div>
      </div>

      <div className="flex flex-col">
        <div className="-my-2 overflow-x-auto sm:-mx-6 lg:-mx-8">
          <div className="py-2 align-middle inline-block min-w-full sm:px-6 lg:px-8">
            <div className="overflow-hidden border-gray-200 sm:rounded-lg">
              <table className="min-w-full divide-y divide-gray-300">
                <thead className="">
                  <tr>
                    <th
                      scope="col"
                      className="px-6 py-3 text-left text-xs font-bold text-gray-700 uppercase tracking-wider">
                      Sciper
                    </th>
                    <th
                      scope="col"
                      className="px-6 py-3 text-left text-xs font-bold text-gray-700 uppercase tracking-wider">
                      {t('role')}
                    </th>
                    <th scope="col" className="relative px-6 py-3">
                      <span className="sr-only">{t('edit')}</span>
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {scipersToDisplay.map((user) => (
                    <tr key={user.id}>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                        {user.sciper}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {user.role}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                        <div
                          className="cursor-pointer text-indigo-600 hover:text-indigo-900"
                          onClick={() => handleDelete(user.sciper)}>
                          {t('delete')}
                        </div>
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
                      {partitionArray(users, SCIPERS_PER_PAGE).length}
                    </span>{' '}
                    {t('of')} <span className="font-medium">{users.length}</span> {t('results')}
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
                    disabled={partitionArray(users, SCIPERS_PER_PAGE).length <= pageIndex + 1}
                    onClick={handleNext}
                    className="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
                    {t('next')}
                  </button>
                </div>
              </nav>
            </div>
          </div>
        </div>
      </div>
    </div>
  ) : (
    <Loading />
  );
};
export default Admin;
