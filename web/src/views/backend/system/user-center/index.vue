<!-- 个人中心页面 -->
<template>
  <div class="w-full h-full p-0 bg-transparent border-none shadow-none">
    <div class="relative flex-b mt-2.5 max-md:block max-md:mt-1">
      <div class="w-112 mr-5 max-md:w-full max-md:mr-0">
        <div class="art-card-sm relative p-9 pb-6 overflow-hidden text-center">
          <img class="absolute top-0 left-0 w-full h-50 object-cover" src="@imgs/user/bg.webp" />
          <!-- 头像（点击上传） -->
          <ElUpload
            class="relative z-10 mt-30 mx-auto w-20 h-20"
            :show-file-list="false"
            accept="image/*"
            :http-request="handleAvatarUpload"
          >
            <ElTooltip content="点击更换头像" placement="bottom">
              <img
                class="w-20 h-20 object-cover border-2 border-white rounded-full cursor-pointer hover:opacity-80 transition-opacity"
                :src="avatarDisplay"
              />
            </ElTooltip>
          </ElUpload>
          <h2 class="mt-5 text-xl font-normal">{{ profileData.nickname || profileData.username }}</h2>
          <p class="mt-5 text-sm">{{ profileData.remark || '暂无个人简介' }}</p>

          <div class="w-75 mx-auto mt-7.5 text-left">
            <div class="mt-2.5">
              <ArtSvgIcon icon="ri:mail-line" class="text-g-700" />
              <span class="ml-2 text-sm">{{ profileData.email || '暂无邮箱' }}</span>
            </div>
            <div class="mt-2.5">
              <ArtSvgIcon icon="ri:user-3-line" class="text-g-700" />
              <span class="ml-2 text-sm">{{ postLabel }}</span>
            </div>
            <div class="mt-2.5">
              <ArtSvgIcon icon="ri:map-pin-line" class="text-g-700" />
              <span class="ml-2 text-sm">{{ profileData.address || '暂无地址' }}</span>
            </div>
            <div class="mt-2.5">
              <ArtSvgIcon icon="ri:dribbble-fill" class="text-g-700" />
              <span class="ml-2 text-sm">{{ profileData.deptFullPath || '暂无部门' }}</span>
            </div>
          </div>

          <div class="mt-10">
            <h3 class="text-sm font-medium">标签</h3>
            <div class="flex flex-wrap justify-center mt-3.5">
              <div
                v-for="item in lableList"
                :key="item"
                class="py-1 px-1.5 mr-2.5 mb-2.5 text-xs border border-g-300 rounded"
              >
                {{ item }}
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="flex-1 overflow-hidden max-md:w-full max-md:mt-3.5">
        <div class="art-card-sm">
          <h1 class="p-4 text-xl font-normal border-b border-g-300">基本设置</h1>

          <ElForm
            :model="form"
            class="box-border p-5 [&>.el-row_.el-form-item]:w-[calc(50%-10px)] [&>.el-row_.el-input]:w-full [&>.el-row_.el-select]:w-full"
            ref="ruleFormRef"
            :rules="rules"
            label-width="86px"
            label-position="top"
          >
            <ElRow>
              <ElFormItem label="用户名" prop="username">
                <ElInput v-model="profileData.username" disabled />
              </ElFormItem>
              <ElFormItem label="性别" prop="sex" class="ml-5">
                <ElSelect v-model="form.sex" placeholder="Select" :disabled="!isEdit">
                  <ElOption
                    v-for="item in options"
                    :key="item.value"
                    :label="item.label"
                    :value="item.value"
                  />
                </ElSelect>
              </ElFormItem>
            </ElRow>

            <ElRow>
              <ElFormItem label="昵称" prop="nikeName">
                <ElInput v-model="form.nikeName" :disabled="!isEdit" />
              </ElFormItem>
              <ElFormItem label="邮箱" prop="email" class="ml-5">
                <ElInput v-model="form.email" :disabled="!isEdit" />
              </ElFormItem>
            </ElRow>

            <ElRow>
              <ElFormItem label="手机" prop="mobile">
                <ElInput v-model="form.mobile" :disabled="!isEdit" />
              </ElFormItem>
              <ElFormItem label="地址" prop="address" class="ml-5">
                <ElInput v-model="form.address" :disabled="!isEdit" />
              </ElFormItem>
            </ElRow>

            <ElFormItem label="个人介绍" prop="des" class="h-32">
              <ElInput type="textarea" :rows="4" v-model="form.des" :disabled="!isEdit" />
            </ElFormItem>

            <div class="flex-c justify-end [&_.el-button]:!w-27.5">
              <ElButton type="primary" class="w-22.5" v-ripple @click="edit">
                {{ isEdit ? '保存' : '编辑' }}
              </ElButton>
            </div>
          </ElForm>
        </div>

        <div class="art-card-sm my-5">
          <h1 class="p-4 text-xl font-normal border-b border-g-300">更改密码</h1>

          <ElForm :model="pwdForm" class="box-border p-5" label-width="86px" label-position="top">
            <ElFormItem label="当前密码" prop="password">
              <ElInput
                v-model="pwdForm.password"
                type="password"
                :disabled="!isEditPwd"
                show-password
              />
            </ElFormItem>

            <ElFormItem label="新密码" prop="newPassword">
              <ElInput
                v-model="pwdForm.newPassword"
                type="password"
                :disabled="!isEditPwd"
                show-password
              />
            </ElFormItem>

            <ElFormItem label="确认新密码" prop="confirmPassword">
              <ElInput
                v-model="pwdForm.confirmPassword"
                type="password"
                :disabled="!isEditPwd"
                show-password
              />
            </ElFormItem>

            <div class="flex-c justify-end [&_.el-button]:!w-27.5">
              <ElButton type="primary" class="w-22.5" v-ripple @click="editPwd">
                {{ isEditPwd ? '保存' : '编辑' }}
              </ElButton>
            </div>
          </ElForm>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { useUserStore } from '@/store/modules/user'
  import { fetchGetUserInfo, fetchUpdateProfile, fetchChangePassword } from '@/api/backend/auth'
  import { adminRequest } from '@/utils/http'
  import { mediaDisplayUrl, ensureMediaPreviewUrl, uploadStoragePath } from '@/utils/media-url'
  import type { FormInstance, FormRules } from 'element-plus'
  import { ElMessage } from 'element-plus'

  defineOptions({ name: 'UserCenter' })

  const userStore = useUserStore()
  const userInfo = computed(() => userStore.getUserInfo)

  const isEdit = ref(false)
  const isEditPwd = ref(false)
  const date = ref('')
  const ruleFormRef = ref<FormInstance>()

  // 真实用户数据
  const profileData = ref<any>({})

  /** 头像显示：有头像用 CDN/可访问 URL，无头像用 DiceBear 字母头像 */
  const avatarDisplay = computed(() => {
    if (profileData.value.avatar) return mediaDisplayUrl(profileData.value.avatar)
    const name = encodeURIComponent(profileData.value.username || profileData.value.nickname || '?')
    return `https://api.dicebear.com/7.x/initials/svg?seed=${name}&backgroundColor=5a8dee,10b981,ff6b6b,ffab00,03c3ec&fontSize=40`
  })

  watch(
    () => profileData.value.avatar,
    (path) => {
      if (path && !/^https?:\/\//i.test(path)) void ensureMediaPreviewUrl(path)
    },
    { immediate: true }
  )

  /** 上传头像 */
  const handleAvatarUpload = async (options: any) => {
    const formData = new FormData()
    formData.append('file', options.file)
    try {
      const res: any = await adminRequest.post({ url: '/upload/file', data: formData })
      const path = uploadStoragePath(res)
      if (!path) {
        ElMessage.error('上传失败')
        return
      }
      // 保存头像到个人资料（入库 path，展示走 CDN 解析）
      await fetchUpdateProfile({
        nickname: profileData.value.nickname || profileData.value.username,
        avatar: path,
        email: profileData.value.email || '',
        gender: profileData.value.gender || 0,
      })
      profileData.value.avatar = path
      ElMessage.success('头像更新成功')
      // 同步到 store
      await loadProfile()
    } catch {
      ElMessage.error('头像上传失败')
    }
  }

  // 性别标签
  const genderLabel = computed(() => {
    const map: Record<number, string> = { 0: '保密', 1: '男', 2: '女' }
    return map[profileData.value.gender] || '保密'
  })

  // 岗位标签
  const postLabel = computed(() => {
    if (!profileData.value.postNames || profileData.value.postNames.length === 0) {
      return '暂无岗位'
    }
    return profileData.value.postNames.join(', ')
  })

  // 角色标签
  const rolesLabel = computed(() => {
    if (!profileData.value.roles || profileData.value.roles.length === 0) {
      return '暂无角色'
    }
    return profileData.value.roles.join(', ')
  })

  /**
   * 用户信息表单
   */
  const form = reactive({
    realName: '',
    nikeName: '',
    email: '',
    mobile: '',
    address: '',
    sex: '0',
    des: ''
  })

  /**
   * 密码修改表单
   */
  const pwdForm = reactive({
    password: '',
    newPassword: '',
    confirmPassword: ''
  })

  /**
   * 表单验证规则
   */
  const rules = reactive<FormRules>({
    realName: [
      { required: true, message: '请输入姓名', trigger: 'blur' },
      { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
    ],
    nikeName: [
      { required: true, message: '请输入昵称', trigger: 'blur' },
      { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
    ],
    email: [{ required: true, message: '请输入邮箱', trigger: 'blur' }],
    mobile: [{ required: true, message: '请输入手机号码', trigger: 'blur' }],
    address: [{ required: true, message: '请输入地址', trigger: 'blur' }],
    sex: [{ required: true, message: '请选择性别', trigger: 'blur' }]
  })

  /**
   * 性别选项
   */
  const options = [
    { value: '0', label: '保密' },
    { value: '1', label: '男' },
    { value: '2', label: '女' }
  ]

  /**
   * 用户标签列表
   */
  const lableList: Array<string> = ['专注设计', '很有想法', '辣~', '大长腿', '川妹子', '海纳百川']

  onMounted(async () => {
    getDate()
    await loadProfile()
  })

  /**
   * 加载用户资料
   */
  const loadProfile = async () => {
    try {
      const res: any = await fetchGetUserInfo()
      profileData.value = res
      // 同步到表单
      form.realName = res.realName || ''
      form.nikeName = res.nickname || ''
      form.email = res.email || ''
      form.mobile = res.mobile || ''
      form.address = res.address || ''
      form.sex = String(res.gender || 0)
      form.des = res.remark || ''
    } catch (error) {
      console.error('加载用户信息失败:', error)
    }
  }

  /**
   * 根据当前时间获取问候语
   */
  const getDate = () => {
    const h = new Date().getHours()

    if (h >= 6 && h < 9) date.value = '早上好'
    else if (h >= 9 && h < 11) date.value = '上午好'
    else if (h >= 11 && h < 13) date.value = '中午好'
    else if (h >= 13 && h < 18) date.value = '下午好'
    else if (h >= 18 && h < 24) date.value = '晚上好'
    else date.value = '很晚了，早点睡'
  }

  /**
   * 切换用户信息编辑状态/保存用户信息
   */
  const edit = async () => {
    if (isEdit.value) {
      // 当前是编辑状态，点击保存
      try {
        await ruleFormRef.value?.validate()
        await fetchUpdateProfile({
          nickname: form.nikeName,
          realName: form.realName,
          avatar: profileData.value.avatar || '',
          email: form.email,
          mobile: form.mobile,
          address: form.address,
          gender: parseInt(form.sex),
          remark: form.des
        })
        ElMessage.success('保存成功')
        isEdit.value = false
        // 重新加载用户信息
        await loadProfile()
      } catch (error) {
        console.error('保存失败:', error)
        if (error !== 'cancel') {
          ElMessage.error('保存失败')
        }
      }
    } else {
      // 当前是查看状态，切换到编辑状态
      isEdit.value = true
    }
  }

  /**
   * 切换密码编辑状态/保存新密码
   */
  const editPwd = async () => {
    if (isEditPwd.value) {
      // 当前是编辑状态，点击保存
      if (!pwdForm.password || !pwdForm.newPassword || !pwdForm.confirmPassword) {
        ElMessage.error('请填写完整的密码信息')
        return
      }
      if (pwdForm.newPassword !== pwdForm.confirmPassword) {
        ElMessage.error('两次密码输入不一致')
        return
      }
      if (pwdForm.newPassword.length < 6 || pwdForm.newPassword.length > 20) {
        ElMessage.error('新密码长度应为6-20个字符')
        return
      }
      try {
        await fetchChangePassword({
          oldPassword: pwdForm.password,
          newPassword: pwdForm.newPassword,
          confirmPassword: pwdForm.confirmPassword
        })
        ElMessage.success('密码修改成功')
        isEditPwd.value = false
        // 清空密码表单
        pwdForm.password = ''
        pwdForm.newPassword = ''
        pwdForm.confirmPassword = ''
      } catch (error) {
        console.error('修改密码失败:', error)
        ElMessage.error('修改密码失败')
      }
    } else {
      // 当前是查看状态，切换到编辑状态
      isEditPwd.value = true
    }
  }
</script>
