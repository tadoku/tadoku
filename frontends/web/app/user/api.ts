import { post } from '../api'

const changePassword = async (payload: {
  oldPassword: string
  newPassword: string
}): Promise<boolean> => {
  const response = await post(`/users/update_password`, {
    body: payload,
    authenticated: true,
  })

  return response.status === 200
}

const UserApi = {
  changePassword,
}

export default UserApi
