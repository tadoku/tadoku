import { post } from '../api'

const changePassword = async (
  userId: number,
  payload: {
    currentPassword: string
    newPassword: string
  },
): Promise<boolean> => {
  const response = await post(`/users/${userId}/password`, {
    body: {
      current_password: payload.currentPassword,
      new_password: payload.newPassword,
    },
  })

  return response.status === 200
}

const updateProfile = async (
  userId: number,
  payload: {
    displayName: string
  },
): Promise<boolean> => {
  const response = await post(`/users/${userId}/profile`, {
    body: {
      display_name: payload.displayName,
    },
  })

  return response.status === 200
}

const UserApi = {
  changePassword,
  updateProfile,
}

export default UserApi
