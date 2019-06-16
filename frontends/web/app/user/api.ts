import { post } from '../api'

const changePassword = async (payload: {
  currentPassword: string
  newPassword: string
}): Promise<boolean> => {
  const response = await post(`/users/update_password`, {
    body: {
      current_password: payload.currentPassword,
      new_password: payload.newPassword,
    },
    authenticated: true,
  })

  return response.status === 200
}

const UserApi = {
  changePassword,
}

export default UserApi
