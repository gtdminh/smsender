import { action, observable } from 'mobx'

export default class MessageModel {
  json
  @observable id
  @observable to
  @observable from
  @observable body
  @observable route
  @observable provider
  @observable status
  @observable created_time // eslint-disable-line

  @action fromJS (object) {
    this.json = object
    this.id = object.id
    this.to = object.to
    this.from = object.from
    this.body = object.body
    this.route = object.route
    this.provider = object.provider
    this.status = object.status
    this.original_message_id = object.original_message_id
    this.created_time = object.created_time

    return this
  }

  toJS () {
    return {
      id: this.id,
      to: this.to,
      from: this.from,
      body: this.body,
      route: this.route,
      provider: this.provider,
      status: this.status,
      original_message_id: this.original_message_id,
      created_time: this.created_time,
      json: this.json
    }
  }

  static new (object) {
    return (new MessageModel()).fromJS(object)
  }
}
