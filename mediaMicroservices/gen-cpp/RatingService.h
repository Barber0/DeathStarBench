/**
 * Autogenerated by Thrift Compiler (0.12.0)
 *
 * DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING
 *  @generated
 */
#ifndef RatingService_H
#define RatingService_H

#include <thrift/TDispatchProcessor.h>
#include <thrift/async/TConcurrentClientSyncInfo.h>
#include "media_service_types.h"

namespace media_service {

#ifdef _MSC_VER
  #pragma warning( push )
  #pragma warning (disable : 4250 ) //inheriting methods via dominance 
#endif

class RatingServiceIf {
 public:
  virtual ~RatingServiceIf() {}
  virtual void UploadRating(const int64_t req_id, const std::string& movie_id, const int32_t rating, const std::map<std::string, std::string> & carrier) = 0;
};

class RatingServiceIfFactory {
 public:
  typedef RatingServiceIf Handler;

  virtual ~RatingServiceIfFactory() {}

  virtual RatingServiceIf* getHandler(const ::apache::thrift::TConnectionInfo& connInfo) = 0;
  virtual void releaseHandler(RatingServiceIf* /* handler */) = 0;
};

class RatingServiceIfSingletonFactory : virtual public RatingServiceIfFactory {
 public:
  RatingServiceIfSingletonFactory(const ::apache::thrift::stdcxx::shared_ptr<RatingServiceIf>& iface) : iface_(iface) {}
  virtual ~RatingServiceIfSingletonFactory() {}

  virtual RatingServiceIf* getHandler(const ::apache::thrift::TConnectionInfo&) {
    return iface_.get();
  }
  virtual void releaseHandler(RatingServiceIf* /* handler */) {}

 protected:
  ::apache::thrift::stdcxx::shared_ptr<RatingServiceIf> iface_;
};

class RatingServiceNull : virtual public RatingServiceIf {
 public:
  virtual ~RatingServiceNull() {}
  void UploadRating(const int64_t /* req_id */, const std::string& /* movie_id */, const int32_t /* rating */, const std::map<std::string, std::string> & /* carrier */) {
    return;
  }
};

typedef struct _RatingService_UploadRating_args__isset {
  _RatingService_UploadRating_args__isset() : req_id(false), movie_id(false), rating(false), carrier(false) {}
  bool req_id :1;
  bool movie_id :1;
  bool rating :1;
  bool carrier :1;
} _RatingService_UploadRating_args__isset;

class RatingService_UploadRating_args {
 public:

  RatingService_UploadRating_args(const RatingService_UploadRating_args&);
  RatingService_UploadRating_args& operator=(const RatingService_UploadRating_args&);
  RatingService_UploadRating_args() : req_id(0), movie_id(), rating(0) {
  }

  virtual ~RatingService_UploadRating_args() throw();
  int64_t req_id;
  std::string movie_id;
  int32_t rating;
  std::map<std::string, std::string>  carrier;

  _RatingService_UploadRating_args__isset __isset;

  void __set_req_id(const int64_t val);

  void __set_movie_id(const std::string& val);

  void __set_rating(const int32_t val);

  void __set_carrier(const std::map<std::string, std::string> & val);

  bool operator == (const RatingService_UploadRating_args & rhs) const
  {
    if (!(req_id == rhs.req_id))
      return false;
    if (!(movie_id == rhs.movie_id))
      return false;
    if (!(rating == rhs.rating))
      return false;
    if (!(carrier == rhs.carrier))
      return false;
    return true;
  }
  bool operator != (const RatingService_UploadRating_args &rhs) const {
    return !(*this == rhs);
  }

  bool operator < (const RatingService_UploadRating_args & ) const;

  uint32_t read(::apache::thrift::protocol::TProtocol* iprot);
  uint32_t write(::apache::thrift::protocol::TProtocol* oprot) const;

};


class RatingService_UploadRating_pargs {
 public:


  virtual ~RatingService_UploadRating_pargs() throw();
  const int64_t* req_id;
  const std::string* movie_id;
  const int32_t* rating;
  const std::map<std::string, std::string> * carrier;

  uint32_t write(::apache::thrift::protocol::TProtocol* oprot) const;

};

typedef struct _RatingService_UploadRating_result__isset {
  _RatingService_UploadRating_result__isset() : se(false) {}
  bool se :1;
} _RatingService_UploadRating_result__isset;

class RatingService_UploadRating_result {
 public:

  RatingService_UploadRating_result(const RatingService_UploadRating_result&);
  RatingService_UploadRating_result& operator=(const RatingService_UploadRating_result&);
  RatingService_UploadRating_result() {
  }

  virtual ~RatingService_UploadRating_result() throw();
  ServiceException se;

  _RatingService_UploadRating_result__isset __isset;

  void __set_se(const ServiceException& val);

  bool operator == (const RatingService_UploadRating_result & rhs) const
  {
    if (!(se == rhs.se))
      return false;
    return true;
  }
  bool operator != (const RatingService_UploadRating_result &rhs) const {
    return !(*this == rhs);
  }

  bool operator < (const RatingService_UploadRating_result & ) const;

  uint32_t read(::apache::thrift::protocol::TProtocol* iprot);
  uint32_t write(::apache::thrift::protocol::TProtocol* oprot) const;

};

typedef struct _RatingService_UploadRating_presult__isset {
  _RatingService_UploadRating_presult__isset() : se(false) {}
  bool se :1;
} _RatingService_UploadRating_presult__isset;

class RatingService_UploadRating_presult {
 public:


  virtual ~RatingService_UploadRating_presult() throw();
  ServiceException se;

  _RatingService_UploadRating_presult__isset __isset;

  uint32_t read(::apache::thrift::protocol::TProtocol* iprot);

};

class RatingServiceClient : virtual public RatingServiceIf {
 public:
  RatingServiceClient(apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> prot) {
    setProtocol(prot);
  }
  RatingServiceClient(apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> iprot, apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> oprot) {
    setProtocol(iprot,oprot);
  }
 private:
  void setProtocol(apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> prot) {
  setProtocol(prot,prot);
  }
  void setProtocol(apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> iprot, apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> oprot) {
    piprot_=iprot;
    poprot_=oprot;
    iprot_ = iprot.get();
    oprot_ = oprot.get();
  }
 public:
  apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> getInputProtocol() {
    return piprot_;
  }
  apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> getOutputProtocol() {
    return poprot_;
  }
  void UploadRating(const int64_t req_id, const std::string& movie_id, const int32_t rating, const std::map<std::string, std::string> & carrier);
  void send_UploadRating(const int64_t req_id, const std::string& movie_id, const int32_t rating, const std::map<std::string, std::string> & carrier);
  void recv_UploadRating();
 protected:
  apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> piprot_;
  apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> poprot_;
  ::apache::thrift::protocol::TProtocol* iprot_;
  ::apache::thrift::protocol::TProtocol* oprot_;
};

class RatingServiceProcessor : public ::apache::thrift::TDispatchProcessor {
 protected:
  ::apache::thrift::stdcxx::shared_ptr<RatingServiceIf> iface_;
  virtual bool dispatchCall(::apache::thrift::protocol::TProtocol* iprot, ::apache::thrift::protocol::TProtocol* oprot, const std::string& fname, int32_t seqid, void* callContext);
 private:
  typedef  void (RatingServiceProcessor::*ProcessFunction)(int32_t, ::apache::thrift::protocol::TProtocol*, ::apache::thrift::protocol::TProtocol*, void*);
  typedef std::map<std::string, ProcessFunction> ProcessMap;
  ProcessMap processMap_;
  void process_UploadRating(int32_t seqid, ::apache::thrift::protocol::TProtocol* iprot, ::apache::thrift::protocol::TProtocol* oprot, void* callContext);
 public:
  RatingServiceProcessor(::apache::thrift::stdcxx::shared_ptr<RatingServiceIf> iface) :
    iface_(iface) {
    processMap_["UploadRating"] = &RatingServiceProcessor::process_UploadRating;
  }

  virtual ~RatingServiceProcessor() {}
};

class RatingServiceProcessorFactory : public ::apache::thrift::TProcessorFactory {
 public:
  RatingServiceProcessorFactory(const ::apache::thrift::stdcxx::shared_ptr< RatingServiceIfFactory >& handlerFactory) :
      handlerFactory_(handlerFactory) {}

  ::apache::thrift::stdcxx::shared_ptr< ::apache::thrift::TProcessor > getProcessor(const ::apache::thrift::TConnectionInfo& connInfo);

 protected:
  ::apache::thrift::stdcxx::shared_ptr< RatingServiceIfFactory > handlerFactory_;
};

class RatingServiceMultiface : virtual public RatingServiceIf {
 public:
  RatingServiceMultiface(std::vector<apache::thrift::stdcxx::shared_ptr<RatingServiceIf> >& ifaces) : ifaces_(ifaces) {
  }
  virtual ~RatingServiceMultiface() {}
 protected:
  std::vector<apache::thrift::stdcxx::shared_ptr<RatingServiceIf> > ifaces_;
  RatingServiceMultiface() {}
  void add(::apache::thrift::stdcxx::shared_ptr<RatingServiceIf> iface) {
    ifaces_.push_back(iface);
  }
 public:
  void UploadRating(const int64_t req_id, const std::string& movie_id, const int32_t rating, const std::map<std::string, std::string> & carrier) {
    size_t sz = ifaces_.size();
    size_t i = 0;
    for (; i < (sz - 1); ++i) {
      ifaces_[i]->UploadRating(req_id, movie_id, rating, carrier);
    }
    ifaces_[i]->UploadRating(req_id, movie_id, rating, carrier);
  }

};

// The 'concurrent' client is a thread safe client that correctly handles
// out of order responses.  It is slower than the regular client, so should
// only be used when you need to share a connection among multiple threads
class RatingServiceConcurrentClient : virtual public RatingServiceIf {
 public:
  RatingServiceConcurrentClient(apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> prot) {
    setProtocol(prot);
  }
  RatingServiceConcurrentClient(apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> iprot, apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> oprot) {
    setProtocol(iprot,oprot);
  }
 private:
  void setProtocol(apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> prot) {
  setProtocol(prot,prot);
  }
  void setProtocol(apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> iprot, apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> oprot) {
    piprot_=iprot;
    poprot_=oprot;
    iprot_ = iprot.get();
    oprot_ = oprot.get();
  }
 public:
  apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> getInputProtocol() {
    return piprot_;
  }
  apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> getOutputProtocol() {
    return poprot_;
  }
  void UploadRating(const int64_t req_id, const std::string& movie_id, const int32_t rating, const std::map<std::string, std::string> & carrier);
  int32_t send_UploadRating(const int64_t req_id, const std::string& movie_id, const int32_t rating, const std::map<std::string, std::string> & carrier);
  void recv_UploadRating(const int32_t seqid);
 protected:
  apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> piprot_;
  apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> poprot_;
  ::apache::thrift::protocol::TProtocol* iprot_;
  ::apache::thrift::protocol::TProtocol* oprot_;
  ::apache::thrift::async::TConcurrentClientSyncInfo sync_;
};

#ifdef _MSC_VER
  #pragma warning( pop )
#endif

} // namespace

#endif
