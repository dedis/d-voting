import { FC, useEffect, useState } from 'react';

import AddAdminUserModal from 'components/modal/AddAdminUserModal';
import { useTranslation } from 'react-i18next';
import RemoveAdminUserModal from 'components/modal/RemoveAdminUserModal';

import { User } from 'types/userRole';

const SCIPERS_PER_PAGE = 5;

type AdminTableProps = {
  users: User[];
  setUsers: (users: User[]) => void;
};

const AdminTable: FC<AdminTableProps> = ({ users, setUsers }) => {
  const { t } = useTranslation();

  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [newUserOpen, setNewUserOpen] = useState(false);
  const [scipersToDisplay, setScipersToDisplay] = useState([]);
  const [sciperToDelete, setSciperToDelete] = useState(0);
  const [pageIndex, setPageIndex] = useState(0);

  const openModal = () => setNewUserOpen(true);

  const partitionArray = (array: User[], size: number) =>
    array.map((_v, i) => (i % size === 0 ? array.slice(i, i + size) : null)).filter((v) => v);

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

  const handleAddRoleUser = (user: User): void => {
    setUsers([...users, user]);
  };

  const handleRemoveRoleUser = (): void => {
    setUsers(users.filter((user) => user.sciper !== sciperToDelete.toString()));
  };

  return (
    <div>
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
      <div className="flex flex-col sm:flex-row sm:items-center justify-between mb-4 py-6 pl-2">
        <div>
          <div className="font-bold uppercase text-lg text-gray-700">{t('roles')}</div>

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

      <div className="relative divide-y overflow-x-auto shadow-md sm:rounded-lg mt-2">
        <table className="w-full text-sm text-left text-gray-500">
          <thead className="text-xs text-gray-700 uppercase bg-gray-50">
            <tr>
              <th scope="col" className="px-6 py-3">
                Sciper
              </th>
              <th scope="col" className="px-6 py-3">
                {t('role')}
              </th>
              <th scope="col" className="px-6 py-3">
                <span className="sr-only">{t('edit')}</span>
              </th>
            </tr>
          </thead>
          <tbody>
            {scipersToDisplay.map((user) => (
              <tr key={user.id} className="bg-white border-b hover:bg-gray-50">
                <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                  {user.sciper}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{user.role}</td>
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
              <span className="font-medium">{partitionArray(users, SCIPERS_PER_PAGE).length}</span>{' '}
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

      {/*<div className="flex flex-col">
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
                  </div> */}
    </div>
  );
};
export default AdminTable;
