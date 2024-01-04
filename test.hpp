/************************************************************************CPY11*/
/*   Copyright Mentor Graphics Corporation 2013  All Rights Reserved.    CPY12*/
/*                                                                       CPY13*/
/*   THIS WORK CONTAINS TRADE SECRET AND PROPRIETARY INFORMATION         CPY14*/
/*   WHICH IS THE PROPERTY OF MENTOR GRAPHICS CORPORATION OR ITS         CPY15*/
/*   LICENSORS AND IS SUBJECT TO LICENSE TERMS.                          CPY16*/
/************************************************************************CPY17*/
///////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////
//
//  MODULE:        hen_utils.hpp
//
//  AUTHOR:        Levon Manukyan
//
//  DESCRIPTION:
//
//  NOTES:
//
//
///////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////

#ifndef INCLUDED_D2S__PONTE__HEN__HEN_UTILS_HPP
#define INCLUDED_D2S__PONTE__HEN__HEN_UTILS_HPP

#include <ponte/hen/hen.hpp>
#include <ponte/hen/netinst.hpp>

#include <ponte/graphlib/graphwithproperties.hpp>

#include <ponte/base/serialization.hpp>

#include <ponte/base/string_utils.hpp>

namespace HEN
{

// TODO: split into smaller headers
class Utils
{
public: // types
    typedef GraphLib::IReadOnlyGraph::VertexCount VertexCount;
    // Path in the design hierarchy tree going from root cell to the target
    // cell instance
    class HierPath
    {
    public:
        HierPath(CellRef rc)
            : rootCell_(rc)
            , instPath_()
        { }

        // get root (top) cell of the design hierarchy
        CellRef rootCell() const
        {
            return rootCell_;
        }

        // get the depth of the target instance
        size_t depth() const
        {
            return instPath_.size();
        }

        // push another level in the hierarchy path
        void pushLevel(CellInstanceRef cellInstRef)
        {
            instPath_.push_back(cellInstRef);
        }

        // pop the last level from the hierarchy path. Depth must be at least 1
        void popLevel()
        {
            va_check(!instPath_.empty(), "");
            instPath_.pop_back();
        }

        // get the level of interest from the hierarchy path (1, 2, ...,
        // depth())
        CellInstanceRef getLevel(size_t level) const
        {
                va_check(level > 0 && level - 1 < instPath_.size(), "");
                return instPath_[level - 1];
        }

    private:
        typedef Vector<CellInstanceRef> CellInstances;

        CellRef         rootCell_;
        CellInstances   instPath_;
    };

    struct DeepNodeRef
    {
        DeepNodeRef(CellRef rootCell, NetRef n, NodeRef gn)
            : hierPath(rootCell)
            , net(n)
            , node(gn)
        { }

        HierPath        hierPath;
        NetRef          net;
        NodeRef         node;
    };

    typedef BlockVector<CellRef>    CellCollection;

    typedef Iterators::PolymorphIteratorManager<const NetInst&> FlatNetIteratorManager;
    typedef Iterators::ForwardIterator<FlatNetIteratorManager> FlatNetIterator;

    class IDeepSubCellFilter
    {
    public:
        enum Status
        {
            REJECT = 0,

            // If this bit is set, then the current sub cell
            // must be returned
            ACCEPT = 1,

            // If this bit is set, then the sub cells of the
            // the current sub cell must be expanded, otherwise
            // its entire subtree can be skipped
            EXPAND = 2
        };

        virtual ~IDeepSubCellFilter() {}
        Status status(const CellInstData& cellInstData) const
        {
            return statusImpl(cellInstData);
        }

    private:
        virtual Status statusImpl(const CellInstData& cellInstData) const = 0;
    };

    class IFlatNetFilter : public IDeepSubCellFilter
    {
    public:
        bool accept(const NetInst& netInfo) const
        {
            return acceptImpl(netInfo);
        }

    private:
        virtual bool acceptImpl(const NetInst&) const = 0;
    };

    class HierNetGraphProxy;
    
    // FIXME: the name of this class does not reflect that
    // FIXME: it is usable only with CellInstData
    class HierPathObtainer
    {
    public:
        explicit HierPathObtainer(const HENInterface& hen);

        std::string operator()(const CellInstData& cellInstData) const
        {
            return getHierPath(cellInstData);
        }

        std::string getHierPath(const CellInstData& cellInstData) const;

    private:
        typedef IHEN::CellInstProperties CellInstPropMgr;
        typedef CellInstPropMgr::PropertyRef<std::string> CellInstNamePropRef;
        const CellInstNamePropRef cellInstNamePropRef_;
    };

    class NetMapper
    {
    public:
        NetMapper();

        void mapNodeRef(ulong nodeNumber, NodeRef nodeRef);
        NodeRef getNodeRef(ulong nodeNumber) const;
        ulong getNodeNumber(NodeRef nodeRef) const;

        void mapPin(ulong pinNumber, const NodeRef& pinNodeRef);
        NodeRef getPinNodeRef(ulong pinNumber) const;

    private:
        // Members
        typedef Ponte::Map<ulong, NodeRef> PinMap;
        PinMap pinMap_;

        //Copy-protection
        NetMapper(const NetMapper&);
        NetMapper& operator =(const NetMapper&);
    };

    class CellMapper
    {
    public:
        CellMapper();

        void clearNetsMap();
        void clear();

        NetMapper& getNetMapper(ulong n) const;

        void mapNetRef(ulong netNumber, NetRef netRef);
        NetRef getNetRef(ulong netNumber) const;

        void mapCellInstRef(ulong placementNumber, CellInstanceRef cellInstRef);
        CellInstanceRef getCellInstRef(ulong placementNumber) const;

        void mapDevInstRef(ulong devInstId, CellInstanceRef devInstRef);
        CellInstanceRef getDevInstRef(ulong devInstId) const;

    private:
        typedef Ponte::Map<ulong, CellInstanceRef> CellInstRefMap;
        typedef Ponte::Map<ulong, CellInstanceRef> DevInstRefMap;
        typedef Ponte::Map<ulong, NetRef> NetRefMap;
        typedef Ponte::Map<ulong, boost::shared_ptr<NetMapper> > NetRefMapper;
        mutable NetRefMapper netMappers_;
        NetRefMap netRefMap_;
        CellInstRefMap cellInstRefMap_;
        DevInstRefMap devInstRefMap_;

        ///Copy-protection
        CellMapper(const CellMapper&);
        CellMapper& operator =(const CellMapper&);
    };

private: // types
    template<class RefType>
    class HierNetGraphProxyProperties : public ReadOnlyPropertyTableBased<ReadOnlyPropertyAccessorProxy<const IReadOnlyPropertyAccessor<PropertyMgrCfg<RefType, SupportedPropTypes> > > >
    {
    protected:
        typedef PropertyMgrCfg<RefType, SupportedPropTypes> Cfg;
        typedef IReadOnlyPropertyAccessor<Cfg> InterfaceType;
    };

    class HierNetGraphProxyVertexProperties : public  HierNetGraphProxyProperties<NodeRef>
    {
    public:
        explicit HierNetGraphProxyVertexProperties(const HENInterface& hen, const CellInstData& cellInstPropValues, const InterfaceType& p);
    };

    class HierNetGraphProxyEdgeProperties : public  HierNetGraphProxyProperties<EdgeRef>
    {
    public:
        explicit HierNetGraphProxyEdgeProperties(const HENInterface& hen, const CellInstData& cellInstPropValues, const InterfaceType& p);
    };

public: // types
    class HierNetGraphProxy : public GraphLib::ReadOnlyGraphProxy
                            , public GraphLib::VertexPropertyAccessor<HierNetGraphProxyVertexProperties>
                            , public GraphLib::EdgePropertyAccessor<HierNetGraphProxyEdgeProperties>
    {
        typedef PropertyMgrCfg<NodeRef, SupportedPropTypes> GraphVertexPropertiesCfg;
        typedef PropertyMgrCfg<EdgeRef, SupportedPropTypes>   GraphEdgePropertiesCfg;

        typedef IReadOnlyPropertyAccessor<GraphVertexPropertiesCfg> GraphVertexReadOnlyProps;
        typedef IReadOnlyPropertyAccessor<GraphEdgePropertiesCfg> GraphEdgeReadOnlyProps;
        typedef GraphLib::VertexPropertyAccessor<HierNetGraphProxyVertexProperties> VertexProperties;
        typedef GraphLib::EdgePropertyAccessor<HierNetGraphProxyEdgeProperties>   EdgeProperties;

        typedef InterfaceCollection<const IReadOnlyGraph, const GraphVertexReadOnlyProps, const GraphEdgeReadOnlyProps> GraphWithPropertiesInterfaces;
        typedef GraphWithPropertiesInterfaces::PtrType GraphWithPropertiesPtr;

        const GraphWithPropertiesPtr g_;
        const bool mediationRequired_;

    public:
        explicit HierNetGraphProxy(const HENInterface& hen, const NetInst& netInst, GraphWithPropertiesPtr g)
            : ReadOnlyGraphProxy(g.as<IReadOnlyGraph>())
            , VertexProperties(hen, netInst.cellInstData, *(g.as<GraphVertexReadOnlyProps>()))
            , EdgeProperties(hen, netInst.cellInstData, *(g.as<GraphEdgeReadOnlyProps>()))
            , g_(g)
            , mediationRequired_(netInst.cellInstData.depth() > 0)
        {}

        GraphWithPropertiesPtr operator&() const
        {
            return mediationRequired_ ? GraphWithPropertiesPtr(this) : g_;
        }
    };

public: // functions
    static
    FlatNetIterator
    getFlatNets(const HENInterface& hen, CellRef cell, IFlatNetFilter* netFilter);

    static CellRef getTopCell(const HENInterface& hen);
    static CellRef getCellByName(const HENInterface& hen, const std::string& cellName);

    static
    RectangleType
    graphBoundingBox(HENInterface::ConstNetGraphPtr graphPtr);

    // Return an integer-rectangle (almost always the smallest one) strictly
    // containing all the nodes of the given net (i.e. there are no points on
    // the boundaries of the box)
    //
    // NOTE: It is supposed that PG_NODE_X_PROPERTY/PG_NODE_Y_PROPERTY property
    // NOTE: values are properly set for graph nodes
    static
    RectangleType
    getBoundingBox(const HENInterface& hen, const FQNetRef& fqNet);

    // Return an upper bound for the given net graph vertex count
    static
    VertexCount
    getNetGraphVertexCountEstimate(const HENInterface& hen, const FQNetRef& fqNet);

    // Serialize corresponding implementation of HEN
    static void serializeHEN(const HENInterface& hen, ISerialization& s);

    // Deserialize corresponding implementation of HEN
    static std::auto_ptr<HENInterface> deserializeHEN(IDeserialization& d);

    static void serialize(const HENInterface& hen, ISerialization& s);

    static void deserialize(IDeserialization& d, HENInterface& henPtr);

    static std::string getCellInstanceName(const IHEN& hen, const FQCellInstanceRef& inst);

    static FQCellInstanceRef getFQCellInstByName(const IHEN& hen,
                                                 const CellRef parentCell,
                                                 const std::string& subCellName);
};

    std::string getNetName(const IHEN& hen, const FQNetRef& net);
    double getNetCapacitance(const IHEN& hen, const FQNetRef& net);

} // namespace HEN

#endif // INCLUDED_D2S__PONTE__HEN__HEN_UTILS_HPP