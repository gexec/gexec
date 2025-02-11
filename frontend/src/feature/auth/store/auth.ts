import { defineStore } from 'pinia'
import { computed, reactive, unref } from 'vue'
import { useRouter } from 'vue-router'

const AUTH_STORAGE_KEYS = Object.freeze({
  accessToken: 'gexec.auth.access_token',
})

interface ParsedBearerToken {
  admin: boolean
  email: string
  exp: number
  iat: number
  ident: string
  iss: string
  login: string
  name: string
}

interface Token {
  accessToken: string
  expires: number
}

interface User {
  displayName: string
  email: string
  isAdmin: boolean
}

function parseJwt(token: string): ParsedBearerToken {
  const base64Url = token.split('.')[1]
  const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
  const jsonPayload = decodeURIComponent(
    atob(base64)
      .split('')
      .map(function (c) {
        return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2)
      })
      .join('')
  )

  return JSON.parse(jsonPayload) satisfies ParsedBearerToken
}

export const useAuthStore = defineStore('auth', () => {
  const router = useRouter()

  const user = reactive<User>({ displayName: '', email: '', isAdmin: false })
  const token = reactive<Token>({ accessToken: '', expires: 0 })

  const isAuthenticated = computed(() => unref(token).accessToken !== '')

  function signInUser(accessToken: string) {
    const parsedToken = parseJwt(accessToken)

    Object.assign(token, {
      accessToken: accessToken,
      expires: parsedToken.exp,
    })
    Object.assign(user, {
      displayName: parsedToken.name,
      email: parsedToken.email,
      isAdmin: parsedToken.admin,
    })

    localStorage.setItem(AUTH_STORAGE_KEYS.accessToken, accessToken)
  }

  function init() {
    const accessToken = localStorage.getItem(AUTH_STORAGE_KEYS.accessToken)

    if (!accessToken) {
      return
    }

    signInUser(accessToken)
  }

  async function signOutUser() {
    Object.assign(user, { displayName: '', email: '', isAdmin: false })
    Object.assign(token, { accessToken: '', expires: 0 })

    localStorage.removeItem(AUTH_STORAGE_KEYS.accessToken)
    await router.push({ name: 'SignIn' })
  }

  return { user, token, isAuthenticated, signInUser, init, signOutUser }
})
