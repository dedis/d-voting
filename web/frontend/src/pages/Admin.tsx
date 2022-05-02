import { useEffect, useState } from 'react';

import { ENDPOINT_USER_RIGHTS } from 'components/utils/Endpoints';
import { PlusIcon } from '@heroicons/react/outline';

import AddAdminUserModal from 'components/modal/AddAdminUserModal';

const SCIPERS_PER_PAGE = 10;

const Admin = () => {
  const [users, setUsers] = useState([]);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [newusrOpen, setNewusrOpen] = useState(false);
  const [scipersToDisplay, setScipersToDisplay] = useState([]);
  const [pageIndex, setPageIndex] = useState(0);

  const openModal = () => setNewusrOpen(true);

  useEffect(() => {
    if (showDeleteModal) {
      return;
    }

    fetch(ENDPOINT_USER_RIGHTS)
      .then((resp) => {
        const jsonData = resp.json();
        jsonData.then((result) => {
          setUsers(result);
        });
      })
      .catch((error) => {
        console.log(error);
      });
  }, [showDeleteModal]);

  const partitionArray = (array: any[], size: number) =>
    array.map((v, i) => (i % size === 0 ? array.slice(i, i + size) : null)).filter((v) => v);

  useEffect(() => {
    if (users.length) {
      setScipersToDisplay(partitionArray(users, SCIPERS_PER_PAGE)[pageIndex]);
    }
  }, [users, pageIndex]);

  const handleDelete = (sciper: number): void => {
    console.log(sciper);
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

  return (
    <div className="w-[50rem] px-4 py-4">
      <AddAdminUserModal open={newusrOpen} setOpen={setNewusrOpen} />
      <div className="flex justify-between pb-2">
        <div className="mt-3 ml-2">Add/remove roles of users from the admin table</div>
        <button
          onClick={openModal}
          className=" whitespace-nowrap inline-flex mb-2 items-center justify-center px-4 py-2 border border-transparent rounded-md shadow-sm text-base font-medium text-white bg-gray-600 hover:bg-gray-700">
          <PlusIcon className="-ml-1 mr-2 h-4 w-4" aria-hidden="true" />
          Add a user
        </button>
      </div>

      <div className="flex flex-col">
        <div className="-my-2 overflow-x-auto sm:-mx-6 lg:-mx-8">
          <div className="py-2 align-middle inline-block min-w-full sm:px-6 lg:px-8">
            <div className="overflow-hidden border border-gray-200 sm:rounded-lg">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th
                      scope="col"
                      className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Sciper
                    </th>
                    <th
                      scope="col"
                      className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Role
                    </th>
                    <th scope="col" className="relative px-6 py-3">
                      <span className="sr-only">Edit</span>
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
                          Delete
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
                    Showing <span className="font-medium">{pageIndex + 1}</span> to{' '}
                    <span className="font-medium">
                      {partitionArray(users, SCIPERS_PER_PAGE).length}
                    </span>{' '}
                    of <span className="font-medium">{users.length}</span> results
                  </p>
                </div>
                <div className="flex-1 flex justify-between sm:justify-end">
                  <button
                    disabled={pageIndex === 0}
                    onClick={handlePrevious}
                    className="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
                    Previous
                  </button>
                  <button
                    disabled={partitionArray(users, SCIPERS_PER_PAGE).length <= pageIndex + 1}
                    onClick={handleNext}
                    className="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
                    Next
                  </button>
                </div>
              </nav>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
export default Admin;
