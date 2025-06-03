import axios from "axios"

export const base = ""

interface Topic {
    id: number
    created_at: number
    title: string
    mode_id: string
}

interface Post {
    topic_id: number
    floor: number
    updated_at: string
    content: string
}

interface Mode {
    id: number
    name: string
    pub: boolean
}

interface Result<T> {
    code: number
    msg: string
    data?: T
}

const req = axios.create({
    baseURL: base + "/api",
    headers: {
        "Content-Type": "application/json;charset=utf-8"
    },
    withCredentials: false
})

req.interceptors.request.use(config => {
    const token = localStorage.getItem("token")
    if (token) {
        config.headers.Authorization = token
    }
    return config
})

req.interceptors.response.use(response => response.data)

export const getUserId = (): Promise<Result<number>> => {
    return req.get("/up/")
}

export const getCategories = (): Promise<Result<Mode[]>> => {
    return req.get("/cv/")
}

export const getDiscussions = (
    offset?: number
): Promise<Result<Topic[]>> => {
    return req.get("/av/", {
        params: offset != null ? {offset} : undefined
    })
}

export const getDiscussionsByCategory = (
    id: string,
    offset?: number
): Promise<Result<Topic[]>> => {
    return req.get("/cv/" + id, {
        params: offset != null ? {offset} : undefined
    })
}

export const getDiscussion = (
    id: string
): Promise<Result<{
    topic: Topic
    posts: Post[]
}>> => {
    return req.get("/av/" + id)
}

export const userLogin = (
    password: string
): Promise<Result<string>> => {
    return req.post("/auth/login", {
        password
    })
}

export const userChangePassword = (
    password: string
): Promise<Result<void>> => {
    return req.post("/user/update", {
        password
    })
}

export const createCategory = (
    name: string,
    deep: number
): Promise<Result<Mode>> => {
    return req.post("/category/create", {
        name,
        deep
    })
}

export const updateCategory = (
    id: number,
    name: string,
    deep: number
): Promise<Result<Mode>> => {
    return req.post("/category/update", {
        id,
        name,
        deep
    })
}

export const deleteCategory = (
    id: number
): Promise<Result<void>> => {
    return req.post("/category/delete", {
        id
    })
}

export const createDiscussion = (
    title: string,
    mode_id: number,
    content: string
): Promise<Result<Topic>> => {
    return req.post("/discussion/create", {
        title,
        mode_id,
        content
    })
}

export const updateDiscussion = (
    id: number,
    title: string,
    mode_id: number
): Promise<Result<Topic>> => {
    return req.post("/discussion/update", {
        id,
        title,
        mode_id
    })
}

export const deleteDiscussion = (
    id: number
): Promise<Result<void>> => {
    return req.post("/discussion/delete", {
        id
    })
}

export const createComment = (
    topic_id: number,
    content: string
): Promise<Result<Post>> => {
    return req.post("/comment/create", {
        topic_id,
        content
    })
}

export const updateComment = (
    topic_id: number,
    floor: number,
    content: string
): Promise<Result<Post>> => {
    return req.post("/comment/update", {
        topic_id,
        floor,
        content
    })
}

export const deleteComment = (
    topic_id: number,
    floor: number
): Promise<Result<void>> => {
    return req.post("/comment/delete", {
        topic_id,
        floor
    })
}
